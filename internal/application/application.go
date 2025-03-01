package application

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Request struct {
	Expression string `json:"expression"`
}

// Структуры для выражений и задач
type Expression struct {
	ID         string  `json:"id"`
	Expression string  `json:"expression"`
	Status     string  `json:"status"`
	Result     float64 `json:"result,omitempty"`
}

type Task struct {
	ID            string  `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int64   `json:"operation_time"`
}

// Глобальные переменные для хранения выражений и задач
var expressions = make(map[string]*Expression)
var tasks = make(chan Task, 10)

type Config struct {
	Addr string
}

func ConfigFromEnv() *Config {
	config := new(Config)
	config.Addr = os.Getenv("PORT")
	if config.Addr == "" {
		config.Addr = "8080"
	}
	return config
}

type Application struct {
	config *Config
}

func New() *Application {
	return &Application{
		config: ConfigFromEnv(),
	}
}

// Генерация уникальных ID
func generateUniqueID() string {
	return uuid.New().String()
}

// Функции для обработки HTTP запросов
func AddExpressionHandler(w http.ResponseWriter, r *http.Request) {
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid expression", http.StatusUnprocessableEntity)
		return
	}

	// Генерация уникального ID для выражения
	expressionID := generateUniqueID()

	// Создание нового выражения
	expression := &Expression{
		ID:         expressionID,
		Expression: req.Expression,
		Status:     "pending", // Статус пока в ожидании
	}

	// Сохраняем выражение (например, в памяти или базе данных)
	expressions[expressionID] = expression

	// Генерируем задачу для вычислений
	task := Task{
		ID:        expressionID,
		Arg1:      2.0, // Примерные аргументы, возможно потребуется вычислить из строки
		Arg2:      3.0,
		Operation: "+", // Пример операции
	}

	// Отправляем задачу в канал
	tasks <- task

	// Отправляем ответ с ID нового выражения
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

	return http.ListenAndServe(":"+a.config.Addr, r)
}
