package application

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Request – структура входящего запроса с выражением
type Request struct {
	Expression string `json:"expression"`
}

// Expression – структура для хранения выражения и его состояния
type Expression struct {
	ID         string  `json:"id"`
	Expression string  `json:"expression"`
	Status     string  `json:"status"`
	Result     float64 `json:"result,omitempty"`
}

// Task – структура задачи для вычисления
type Task struct {
	ID            string  `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int64   `json:"operation_time"`
}

// Глобальные переменные для хранения выражений и очереди задач
var expressions = make(map[string]*Expression)
var tasks = make(chan Task, 10) // Буферизованный канал для задач

// Config – конфигурация приложения
type Config struct {
	Addr string
}

// ConfigFromEnv – загрузка конфигурации из переменных окружения
func ConfigFromEnv() *Config {
	config := new(Config)
	config.Addr = os.Getenv("PORT")
	if config.Addr == "" {
		config.Addr = "8080"
	}
	return config
}

// Application – основная структура приложения
type Application struct {
	config *Config
}

// New – создание нового экземпляра приложения
func New() *Application {
	return &Application{
		config: ConfigFromEnv(),
	}
}

// generateUniqueID – генерация уникального идентификатора
func generateUniqueID() string {
	return uuid.New().String()
}

// parseExpression – функция для парсинга математического выражения в формате "<number> <operator> <number>"
func parseExpression(expr string) (float64, float64, string, error) {
	parts := strings.Fields(expr)
	if len(parts) != 3 {
		return 0, 0, "", fmt.Errorf("invalid expression format, expected format: <number> <operator> <number>")
	}
	arg1, err1 := strconv.ParseFloat(parts[0], 64)
	arg2, err2 := strconv.ParseFloat(parts[2], 64)
	if err1 != nil || err2 != nil {
		return 0, 0, "", fmt.Errorf("error parsing numbers: %v, %v", err1, err2)
	}
	operator := parts[1]
	if operator != "+" && operator != "-" && operator != "*" && operator != "/" {
		return 0, 0, "", fmt.Errorf("unsupported operator: %s", operator)
	}
	return arg1, arg2, operator, nil
}

// AddExpressionHandler – обработчик POST-запроса для добавления нового выражения
func AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid expression payload", http.StatusBadRequest)
		return
	}

	arg1, arg2, operator, err := parseExpression(req.Expression)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	expressionID := generateUniqueID()

	expr := &Expression{
		ID:         expressionID,
		Expression: req.Expression,
		Status:     "pending",
	}

	// Защищаем доступ к глобальной карте expressions
	expressionsMutex.Lock()
	expressions[expressionID] = expr
	expressionsMutex.Unlock()

	task := Task{
		ID:        expressionID,
		Arg1:      arg1,
		Arg2:      arg2,
		Operation: operator,
	}

	// Отправляем задачу в канал для обработки агентом
	select {
	case tasks <- task:
		log.Printf("Задача с ID %s добавлена в канал", expressionID)
		// Обновляем статус на "processing"
		expressionsMutex.Lock()
		expr.Status = "processing"
		expressionsMutex.Unlock()
	default:
		http.Error(w, "канал задач переполнен", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": expressionID})
}

func GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	var expressionList []Expression
	for _, expr := range expressions {
		expressionList = append(expressionList, *expr)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"expressions": expressionList,
	})
}

func GetExpressionByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	expr, found := expressions[id]
	if !found {
		http.Error(w, "expression not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(expr)
}

func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	task, found := getNextTaskToProcess()
	if !found {
		http.Error(w, "no task available", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

// Логика обработки задач
func getNextTaskToProcess() (Task, bool) {
	select {
	case task := <-tasks:
		return task, true
	default:
		return Task{}, false
	}
}

// Функция для выполнения вычислений
func processTask(task Task) {
	var result float64
	switch task.Operation {
	case "+":
		result = task.Arg1 + task.Arg2
	case "-":
		result = task.Arg1 - task.Arg2
	case "*":
		result = task.Arg1 * task.Arg2
	case "/":
		if task.Arg2 == 0 {
			log.Printf("Ошибка: деление на ноль в задаче с ID %s", task.ID)
			return
		}
		result = task.Arg1 / task.Arg2
	}

	// Проверка на NaN или бесконечность
	if math.IsNaN(result) || math.IsInf(result, 0) {
		log.Printf("Ошибка: результат вычисления для задачи с ID %s некорректен: %v", task.ID, result)
		return
	}

	// Обновляем статус задачи на "completed" и сохраняем результат
	expressionsMutex.Lock()
	expr, found := expressions[task.ID]
	if found {
		expr.Status = "completed"
		expr.Result = result
	}
	expressionsMutex.Unlock()

	log.Printf("Задача с ID %s обработана, результат: %f", task.ID, result)
}

// Запуск агента для обработки задач
func startAgent() {
	for {
		task, found := getNextTaskToProcess()
		if found {
			processTask(task)
		} else {
			log.Println("Задач нет в очереди, агент ожидает...")
			time.Sleep(1 * time.Second) // Пауза, если задач нет
		}
	}
}

// Функция запуска приложения
func (a *Application) RunServer() error {
	r := mux.NewRouter()

	r.HandleFunc("/api/v1/calculate", AddExpressionHandler).Methods("POST")
	r.HandleFunc("/api/v1/expressions", GetExpressionsHandler).Methods("GET")
	r.HandleFunc("/api/v1/expressions/{id}", GetExpressionByIDHandler).Methods("GET")
	r.HandleFunc("/internal/task", GetTaskHandler).Methods("GET")

	go startAgent() // Запуск агента в отдельной горутине

	fmt.Println("Запуск сервера на порту " + a.config.Addr)

	if err := http.ListenAndServe(":"+a.config.Addr, r); err != nil {
		log.Fatal("Ошибка при запуске сервера:", err)
	}
	return http.ListenAndServe(":"+a.config.Addr, r)
}
