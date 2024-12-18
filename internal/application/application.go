package application

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"github.com/Powdersumm/Yandexlmscalcproject/pkg/calculation"
)

type Request struct {
	Expression string `json:"expression"`
}

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

// Функция запуска приложения
// тут будем читать введенную строку и после нажатия ENTER писать результат работы программы на экране
// если пользователь ввел exit - то останаваливаем приложение
func (a *Application) Run() error {
	for {
		// читаем выражение для вычисления из командной строки
		log.Println("input expression")
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Println("failed to read expression from console")
		}
		// убираем пробелы, чтобы оставить только вычислемое выражение
		text = strings.TrimSpace(text)
		// выходим, если ввели команду "exit"
		if text == "exit" {
			log.Println("aplication was successfully closed")
			return nil
		}
		//вычисляем выражение
		result, err := rpn.Calc(text)
		if err != nil {
			log.Println(text, " calculation failed wit error: ", err)
		} else {
			log.Println(text, "=", result)
		}
	}
}

func CalcHandler(w http.ResponseWriter, r *http.Request) {
	request := new(Request)
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := calculation.Calc(request.Expression)
	if err != nil {
		if errors.Is(err, calculation.ErrInvalidExpression) {
			http.Error(w, "ошибочное выражение", http.StatusBadRequest)
			return
		} else if errors.Is(err, calculation.ErrInvalidZero) {
			http.Error(w, "деление на ноль", http.StatusBadRequest)
			return
		} else if errors.Is(err, calculation.ErrInvalidParentheses) {
			http.Error(w, "ошибка с скобками", http.StatusBadRequest)
			return
		} else if errors.Is(err, calculation.ErrInvalidOperand) {
			http.Error(w, "ошибка в операнде", http.StatusBadRequest)
			return
		} else if errors.Is(err, calculation.ErrInvalidValuesCount) {
			http.Error(w, "ошибка в количестве полученных значений", http.StatusBadRequest)
			return
		} else if errors.Is(err, calculation.ErrInvalidCalculation) {
			http.Error(w, "ошибка в посчитанном выражении", http.StatusBadRequest)
			return
		} else {
			http.Error(w, "неизвестная ошибка ", http.StatusInternalServerError)
		}
	} else {
		fmt.Fprintf(w, "result: %f", result)
	}
}

func (a *Application) RunServer() error {
	http.HandleFunc("/", CalcHandler)
	return http.ListenAndServe(a.config.Addr, nil)

}
