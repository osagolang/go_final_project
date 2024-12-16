package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go_final_project/db"
	"go_final_project/utils"
)

// Task структура для задачи
type Task struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// HandleTask обрабатывает запросы API для задач
func HandleTask(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// addTask добавляет задачу в базу данных
func addTask(w http.ResponseWriter, r *http.Request) {
	// Устанавливаем заголовок ответа
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	// Парсим JSON из тела запроса
	var task Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeError(w, "Invalid JSON format")
		return
	}

	// Получаем текущую дату без временной части
	now := time.Now()
	normalizedNow := time.Date(
		now.Year(), now.Month(), now.Day(),
		0, 0, 0, 0, now.Location(),
	)

	// Если дата указана, проверяем её формат
	if task.Date != "" {
		parsedDate, err := time.Parse("20060102", task.Date)
		if err != nil {
			writeError(w, "Invalid date format (expected YYYYMMDD)")
			return
		}

		// Если дата в прошлом или сегодня
		if parsedDate.Before(normalizedNow) || parsedDate.Equal(normalizedNow) {

			if task.Repeat == "" {
				// Если нет правила повторения, подставляем сегодняшнюю дату
				task.Date = normalizedNow.Format("20060102")
			} else if parsedDate.Equal(normalizedNow) && task.Repeat == "d 1" {
				// Если дата сегодня и правило d 1, оставляем сегодняшнюю дату
			} else if parsedDate.Before(normalizedNow) {
				// Если дата в прошлом, вычисляем следующую дату
				task.Date, err = utils.NextDate(normalizedNow, task.Date, task.Repeat)
				if err != nil {
					writeError(w, "Invalid repeat rule")
					return
				}
			}
		}
	} else {
		// Если дата не указана, подставляем сегодняшнюю
		task.Date = normalizedNow.Format("20060102")
	}

	// Проверяем заголовок задачи
	if task.Title == "" {
		writeError(w, "Task title is required")
		return
	}

	// Подключаемся к базе данных
	dbFile := db.GetDatabasePath()
	dbConn, err := sql.Open("sqlite", dbFile)
	if err != nil {
		writeError(w, "Failed to connect to database")
		return
	}
	defer dbConn.Close()

	// Вставляем задачу в базу данных
	res, err := dbConn.Exec(
		"INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)",
		task.Date, task.Title, task.Comment, task.Repeat,
	)
	if err != nil {
		writeError(w, "Failed to add task")
		return
	}

	// Получаем ID добавленной задачи
	id, err := res.LastInsertId()
	if err != nil {
		writeError(w, "Failed to retrieve task ID")
		return
	}

	// Возвращаем ID задачи в формате JSON
	response := map[string]any{"id": id}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		writeError(w, "Failed to encode response")
	}
}

// writeError отправляет сообщение об ошибке в формате JSON
func writeError(w http.ResponseWriter, message string) {
	log.Printf("[ERROR] %s", message)
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{"error": message})
}
