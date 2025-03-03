package application

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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
var tasks = make(chan Task, 10)

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
// Например: "5 * 7"
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
	// Допускаются операторы +, -, *, /
	if operator != "+" && operator != "-" && operator != "*" && operator != "/" {
		return 0, 0, "", fmt.Errorf("unsupported operator: %s", operator)
	}
	return arg1, arg2, operator, nil
}

// AddExpressionHandler – обработчик POST-запроса для добавления нового выражения
func AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	var req Request
	// Декодируем JSON из тела запроса
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid expression payload", http.StatusBadRequest)
		return
	}

	// Парсим выражение, чтобы извлечь операнды и оператор
	arg1, arg2, operator, err := parseExpression(req.Expression)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Генерируем уникальный ID для выражения
	expressionID := generateUniqueID()

	// Создаём новое выражение со статусом "pending"
	expr := &Expression{
		ID:         expressionID,
		Expression: req.Expression,
		Status:     "pending",
	}
	// Сохраняем выражение в памяти
	expressions[expressionID] = expr

	// Формируем задачу для вычислений на основе разобранного выражения
	task := Task{
		ID:        expressionID,
		Arg1:      arg1,
		Arg2:      arg2,
		Operation: operator,
	}

	// Отправляем задачу в канал для дальнейшей обработки агентом
	tasks <- task

	// Отправляем клиенту ответ с ID созданного выражения
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

	// Ищем выражение по ID
	expr, found := expressions[id]
	if !found {
		http.Error(w, "expression not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(expr)
}

func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Ищем задачу для выполнения
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

// Функция запуска приложения
func (a *Application) RunServer() error {
	r := mux.NewRouter()

	// Эндпоинты для оркестратора
	r.HandleFunc("/api/v1/calculate", AddExpressionHandler).Methods("POST")
	r.HandleFunc("/api/v1/expressions", GetExpressionsHandler).Methods("GET")
	r.HandleFunc("/api/v1/expressions/{id}", GetExpressionByIDHandler).Methods("GET")
	r.HandleFunc("/internal/task", GetTaskHandler).Methods("GET")

	fmt.Println("Запуск сервера на порту " + a.config.Addr)

	if err := http.ListenAndServe(":"+a.config.Addr, r); err != nil {
		log.Fatal("Ошибка при запуске сервера:", err)
	}
}
