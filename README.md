<<<<<<< HEAD
# Руководство по запуску и использованию сервера

Этот проект предоставляет сервер для вычисления математических выражений через API. Сервер принимает запросы на выполнение вычислений и возвращает результаты с использованием уникальных идентификаторов для каждой задачи.

Проект размещен на [GitHub](https://github.com/Powdersumm/Yandexlmscalcproject2sprint.git).


### Требования

1. Установленный [Postman](https://www.postman.com/downloads/)
2. Установленный [Git](https://git-scm.com/)


## Установка и запуск сервера

### 1. Установка Go и Postman 
Убедитесь, что у вас установлен [Go](https://golang.org/dl/). Вы можете скачать и установить его, следуя инструкциям на официальном сайте.

### 2. Клонирование репозитория

Для клонирования репозитория выполните следующую команду в терминале:

```bash
git clone https://github.com/Powdersumm/Yandexlmscalcproject2sprint.git
```


### 3. Запуск сервера

1. Откройте терминал и перейдите в директорию проекта:

    ```bash
    cd /путь/к/проекту
    ```

2. Запустите сервер:

    ```bash
    go run cmd/main.go
    ```

3. Сервер будет работать на `localhost:8080` и готов принимать запросы.

---

## Использование через PowerShell

Вы можете использовать **PowerShell** для взаимодействия с сервером с помощью команд `Invoke-RestMethod`.

### 1. Добавление выражения (POST запрос)

Чтобы добавить математическое выражение на сервер для вычислений, используйте **POST-запрос** на адрес `http://localhost:8080/api/v1/calculate`.

### Пример команды в PowerShell для отправки запроса:

```bash
Invoke-RestMethod -Uri http://localhost:8080/api/v1/calculate -Method Post -Body '{"expression": "2341615612424322 * 4"}' -ContentType "application/json"
```

после вы получаете ответ с ID:
id
--
0704a462-9683-4430-abf2-8f05c0b82047

### Для получения ответа нужно использовать команду `Invoke-RestMethod` с параметром `-Method Get` и указать ID:
```bash
Invoke-RestMethod -Uri http://localhost:8080/api/v1/expressions/{ID}]
```
В моем случае:
```bash
Invoke-RestMethod -Uri http://localhost:8080/api/v1/expressions/0704a462-9683-4430-abf2-8f05c0b82047
```


Получаем такой ответ:
id  0704a462-9683-4430-abf2-8f05c0b82047      
expression  2341615612424322 * 4 
status  completed
result  9.3664624e+15




## Использование через Postman:

### Клонируйте репозиторий: Откройте терминал и выполните команду, чтобы клонировать репозиторий с GitHub:
```bash
git clone https://github.com/Powdersumm/Yandexlmscalcproject2sprint.git
```
### Откройте Postman: Если у вас нет Postman, скачайте и установите его с официального сайта.
```bash
https://www.postman.com/downloads/
```

### POST запрос для отправки результата (отправка данных с вычисленным результатом)
Этот запрос выполняется через http.Post в вашем коде для отправки результата вычислений обратно в оркестратор.

Параметры:
```bash
URL: http://localhost:8080/internal/task 
```
(вы можете изменить URL в зависимости от конфигурации вашего сервера).
- **Тело запроса:**
  - Метод: `POST`
  - Тип контента: `application/json`
  - Формат данных в теле запроса: `raw JSON`

![Post запрос на отправку выражения на сервер](https://github.com/Powdersumm/Yandexlmscalcproject2sprint/blob/main/photo_2024-10-06_17-51-11.jpg)



### GET запрос для получения задачи (получение данных с сервера)
Ваш код также использует GET запрос для получения задачи. Вы можете тестировать этот запрос через Postman, чтобы увидеть, как сервер возвращает данные.

Параметры:
```bash
URL: http://localhost:8080/internal/task 
```
 (или другой адрес, на котором ваш сервер обрабатывает GET-запросы).
Метод: `GET`


![Get запрос на получение результата вычесления с сервера](https://github.com/Powdersumm/Yandexlmscalcproject2sprint/blob/main/photo_2025-03-03_18-16-32.jpg)




