package application

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Request – структура входящего запроса с выражением
type Request struct {
	Expression string `json:"expression"`
}

var expressionsMutex = &sync.Mutex{}

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
func parseComplexExpression(expr string) (float64, error) {
	ev, err := govaluate.NewEvaluableExpression(expr)
	if err != nil {
		return 0, fmt.Errorf("ошибка при парсинге выражения: %v", err)
	}
	result, err := ev.Evaluate(nil)
	if err != nil {
		return 0, fmt.Errorf("ошибка при вычислении: %v", err)
	}
	return result.(float64), nil
}

// AddExpressionHandler – обработчик POST-запроса для добавления нового выражения
func AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid expression payload", http.StatusBadRequest)
		return
	}

	// Используем функцию parseComplexExpression для вычисления результата
	result, err := parseComplexExpression(req.Expression)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Генерация уникального ID для выражения
	expressionID := generateUniqueID()

	expr := &Expression{
		ID:         expressionID,
		Expression: req.Expression,
		Status:     "pending",
		Result:     result, // Записываем результат сразу
	}

	// Защищаем доступ к глобальной карте expressions
	expressionsMutex.Lock()
	expressions[expressionID] = expr
	expressionsMutex.Unlock()

	// Возвращаем ответ с ID выражения
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
