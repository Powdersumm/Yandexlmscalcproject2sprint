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
		task, found := getTask()
		if !found {
			log.Println("No task available, waiting...")
			time.Sleep(2 * time.Second)
			continue
		}

		// Запускаем горутину для обработки каждой задачи
		go func(task Task) {
			// Выполняем вычисление задачи
			result, err := performCalculation(task)
			if err != nil {
				log.Println("Error performing calculation:", err)
				return
			}

			// Отправляем результат обратно в оркестратор
			err = sendResult(task.ID, result)
			if err != nil {
				log.Println("Error sending result:", err)
			}

			// Обновляем статус выражения
			expressions[task.ID].Status = "completed"
			expressions[task.ID].Result = result
		}(task)

		time.Sleep(2 * time.Second) // Задержка между задачами
	}
}

func getTask() (Task, error) {
	var task Task
	var err error

	for attempts := 0; attempts < 3; attempts++ {
		resp, err := http.Get("http://localhost:8080/internal/task")
		if err != nil {
			log.Printf("Error sending GET request to /internal/task: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Failed to get task. HTTP status code: %d", resp.StatusCode)
			time.Sleep(2 * time.Second)
			continue
		}

		if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
			log.Printf("Error decoding response body: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("Successfully received task: %v", task)
		return task, nil
	}

	return task, fmt.Errorf("failed to get task after 3 attempts: %v", err)
}

func performCalculation(task Task) (float64, error) {
	// Проверка корректности аргументов (если нужно)
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
		log.Printf("Error marshalling result data: %v\n", err)
		return err
	}

	for attempts := 0; attempts < 3; attempts++ {
		resp, err := http.Post("http://localhost:8080/internal/task", "application/json", bytes.NewBuffer(data))
		if err != nil {
			log.Printf("Error sending result to server: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Failed to send result, received status code: %d\n", resp.StatusCode)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("Successfully sent result for task %s, received status: %d\n", taskID, resp.StatusCode)
		return nil
	}

	return fmt.Errorf("failed to send result after 3 attempts")
}
