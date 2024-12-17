package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"go_final_project/db"
	"go_final_project/models"
	"go_final_project/utils"
)

// HandleTask обрабатывает запросы API для задач
func HandleTask(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTask(w, r)
	case http.MethodGet:
		getTask(w, r)
	case http.MethodPut:
		editTask(w, r)
	case http.MethodDelete:
		deleteTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// addTask добавляет задачу в базу данных
func addTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var task models.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeError(w, "Неверный формат JSON")
		return
	}

	now := utils.NormalizeDate(time.Now())

	if task.Date == "" {
		task.Date = now.Format("20060102")
	} else {
		parsedDate, err := time.Parse("20060102", task.Date)
		if err != nil {
			writeError(w, "Неверный формат даты (ожидается YYYYMMDD)")
			return
		}

		if parsedDate.Before(now) || parsedDate.Equal(now) {
			if task.Repeat == "" {
				task.Date = now.Format("20060102")
			} else {
				task.Date, err = utils.NextDate(now, task.Date, task.Repeat)
				if err != nil {
					writeError(w, "Некорректное правило повторения")
					return
				}
			}
		}
	}

	if task.Title == "" {
		writeError(w, "Не указан заголовок задачи")
		return
	}

	id, err := db.AddTask(task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		writeError(w, "Не удалось добавить задачу")
		return
	}

	response := map[string]any{"id": strconv.FormatInt(id, 10)}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		writeError(w, "Ошибка при формировании ответа")
	}
}

// getTask возвращает данные задачи по идентификатору
func getTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, "Не указан идентификатор задачи")
		return
	}

	taskID, err := strconv.Atoi(id)
	if err != nil {
		writeError(w, "Идентификатор задачи должен быть числом")
		return
	}

	dbFile := db.GetDatabasePath()
	dbConn, err := sql.Open("sqlite", dbFile)
	if err != nil {
		writeError(w, "Не удалось подключиться к базе данных")
		return
	}
	defer dbConn.Close()

	var task models.Task
	row := dbConn.QueryRow(
		"SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?",
		taskID,
	)

	var taskIDStr int64
	err = row.Scan(&taskIDStr, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, "Задача не найдена")
			return
		}
		writeError(w, "Ошибка при получении задачи")
		return
	}

	task.ID = strconv.FormatInt(taskIDStr, 10)

	if err := json.NewEncoder(w).Encode(task); err != nil {
		writeError(w, "Ошибка при формировании ответа")
	}
}

// editTask обновляет параметры задачи
func editTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var task models.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeError(w, "Неверный формат JSON")
		return
	}

	if task.ID == "" {
		writeError(w, "Не указан идентификатор задачи")
		return
	}

	if task.Date != "" {
		if _, err := time.Parse("20060102", task.Date); err != nil {
			writeError(w, "Неверный формат даты (ожидается YYYYMMDD)")
			return
		}
	} else {
		task.Date = utils.NormalizeDate(time.Now()).Format("20060102")
	}

	if task.Title == "" {
		writeError(w, "Заголовок задачи обязателен")
		return
	}

	if task.Repeat != "" {
		now := utils.NormalizeDate(time.Now())
		_, err := utils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeError(w, "Некорректное правило повторения")
			return
		}
	}

	dbFile := db.GetDatabasePath()
	dbConn, err := sql.Open("sqlite", dbFile)
	if err != nil {
		writeError(w, "Не удалось подключиться к базе данных")
		return
	}
	defer dbConn.Close()

	var existingTask models.Task
	err = dbConn.QueryRow(
		"SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?",
		task.ID,
	).Scan(&existingTask.ID, &existingTask.Date, &existingTask.Title, &existingTask.Comment, &existingTask.Repeat)

	if err == sql.ErrNoRows {
		writeError(w, "Задача не найдена")
		return
	} else if err != nil {
		writeError(w, "Ошибка при поиске задачи")
		return
	}

	_, err = dbConn.Exec(
		"UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?",
		task.Date, task.Title, task.Comment, task.Repeat, task.ID,
	)
	if err != nil {
		writeError(w, "Ошибка при обновлении задачи")
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]any{}); err != nil {
		writeError(w, "Ошибка при отправке ответа")
	}
}

// HandleTaskDone завершает задачу
func HandleTaskDone(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, "Не указан идентификатор задачи")
		return
	}

	taskID, err := strconv.Atoi(id)
	if err != nil {
		writeError(w, "Идентификатор задачи должен быть числом")
		return
	}

	dbFile := db.GetDatabasePath()
	dbConn, err := sql.Open("sqlite", dbFile)
	if err != nil {
		writeError(w, "Не удалось подключиться к базе данных")
		return
	}
	defer dbConn.Close()

	var task models.Task
	row := dbConn.QueryRow(
		"SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?",
		taskID,
	)

	var taskIDStr int64
	err = row.Scan(&taskIDStr, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, "Задача не найдена")
			return
		}
		writeError(w, "Ошибка при получении задачи")
		return
	}

	if task.Repeat == "" {
		_, err = dbConn.Exec("DELETE FROM scheduler WHERE id = ?", taskID)
		if err != nil {
			writeError(w, "Не удалось удалить задачу")
			return
		}
	} else {
		now := utils.NormalizeDate(time.Now())
		nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeError(w, "Ошибка при расчёте следующей даты")
			return
		}

		_, err = dbConn.Exec("UPDATE scheduler SET date = ? WHERE id = ?", nextDate, taskID)
		if err != nil {
			writeError(w, "Не удалось обновить задачу")
			return
		}
	}

	if err := json.NewEncoder(w).Encode(map[string]any{}); err != nil {
		writeError(w, "Ошибка при отправке ответа")
	}
}

// deleteTask удаляет задачу по идентификатору
func deleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	id := r.URL.Query().Get("id")
	if id == "" {
		writeError(w, "Не указан идентификатор задачи")
		return
	}

	taskID, err := strconv.Atoi(id)
	if err != nil {
		writeError(w, "Идентификатор задачи должен быть числом")
		return
	}

	dbFile := db.GetDatabasePath()
	dbConn, err := sql.Open("sqlite", dbFile)
	if err != nil {
		writeError(w, "Не удалось подключиться к базе данных")
		return
	}
	defer dbConn.Close()

	_, err = dbConn.Exec("DELETE FROM scheduler WHERE id = ?", taskID)
	if err != nil {
		writeError(w, "Не удалось удалить задачу")
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]any{}); err != nil {
		writeError(w, "Ошибка при отправке ответа")
	}
}

// writeError отправляет сообщение об ошибке в формате JSON
func writeError(w http.ResponseWriter, message string) {
	log.Printf("[ERROR] %s", message)
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(map[string]any{"error": message})
}
