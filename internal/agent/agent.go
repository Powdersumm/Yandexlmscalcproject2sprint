package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Powdersumm/Yandexlmscalcproject2sprint/pkg/calculation"
)

type Task struct {
	ID        string  `json:"id"`
	Arg1      float64 `json:"arg1"`
	Arg2      float64 `json:"arg2"`
	Operation string  `json:"operation"`
}

type Result struct {
	ID     string  `json:"id"`
	Result float64 `json:"result"`
}

func Start() {
	for {
		// Получаем задачу от оркестратора
		task, err := getTask()
		if err != nil {
			log.Println("Error getting task:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Выполняем вычисление задачи
		result, err := performCalculation(task)
		if err != nil {
			log.Println("Error performing calculation:", err)
			continue
		}

		// Отправляем результат обратно в оркестратор
		err = sendResult(task.ID, result)
		if err != nil {
			log.Println("Error sending result:", err)
		}

		time.Sleep(2 * time.Second)
	}
}

func getTask() (Task, error) {
	resp, err := http.Get("http://localhost:8080/internal/task")
	if err != nil {
		return Task{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Task{}, fmt.Errorf("failed to get task, status code: %d", resp.StatusCode)
	}

	var task Task
	if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
		return Task{}, err
	}

	return task, nil
}

func performCalculation(task Task) (float64, error) {
	// Проверка корректности аргументов
	if task.Arg1 == 0 || task.Arg2 == 0 {
		return 0, fmt.Errorf("invalid arguments, task.Arg1 and task.Arg2 must not be zero")
	}

	// Формируем строку выражения для вычислений
	expression := fmt.Sprintf("%f %s %f", task.Arg1, task.Operation, task.Arg2)

	// Используем функцию Calc из пакета calculation для вычислений
	result, err := calculation.Calc(expression)
	if err != nil {
		return 0, fmt.Errorf("error calculating expression: %v", err)
	}

	return result, nil
}

func sendResult(taskID string, result float64) error {
	resultData := Result{
		ID:     taskID,
		Result: result,
	}

	data, err := json.Marshal(resultData)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://localhost:8080/internal/task", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send result, status code: %d", resp.StatusCode)
	}

	return nil
}
