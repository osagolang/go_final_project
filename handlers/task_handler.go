package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"go_final_project/constants"
	"go_final_project/db"
	"go_final_project/models"
	"go_final_project/utils"
)

// HandleTask обрабатывает запросы API для задач
func (h *Handler) HandleTask(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.addTask(w, r)
	case http.MethodGet:
		h.getTask(w, r)
	case http.MethodPut:
		h.editTask(w, r)
	case http.MethodDelete:
		h.deleteTask(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// addTask добавляет задачу
func (h *Handler) addTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	var task models.Task
	err := json.NewDecoder(r.Body).Decode(&task)
	if err != nil {
		writeError(w, "Неверный формат JSON")
		return
	}

	now := utils.NormalizeDate(time.Now())

	if task.Date == "" {
		task.Date = now.Format(constants.DateFormat)
	} else {
		parsedDate, err := time.Parse(constants.DateFormat, task.Date)
		if err != nil {
			writeError(w, "Неверный формат даты (ожидается YYYYMMDD)")
			return
		}

		if parsedDate.Before(now) || parsedDate.Equal(now) {
			if task.Repeat == "" {
				task.Date = now.Format(constants.DateFormat)
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

	id, err := db.AddTask(h.DB, task.Date, task.Title, task.Comment, task.Repeat)
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
func (h *Handler) getTask(w http.ResponseWriter, r *http.Request) {
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

	task, err := db.GetTaskByID(h.DB, taskID)
	if err != nil {
		writeError(w, "Ошибка при получении задачи")
		return
	}

	if err := json.NewEncoder(w).Encode(task); err != nil {
		writeError(w, "Ошибка при формировании ответа")
	}
}

// editTask обновляет параметры задачи
func (h *Handler) editTask(w http.ResponseWriter, r *http.Request) {
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
		if _, err := time.Parse(constants.DateFormat, task.Date); err != nil {
			writeError(w, "Неверный формат даты (ожидается YYYYMMDD)")
			return
		}
	} else {
		task.Date = utils.NormalizeDate(time.Now()).Format(constants.DateFormat)
	}

	if task.Title == "" {
		writeError(w, "Заголовок задачи обязателен")
		return
	}

	rowsAffected, err := db.UpdateTask(h.DB, task)
	if err != nil || rowsAffected == 0 {
		writeError(w, "Задача не найдена или не удалось обновить")
		return
	}

	if err := json.NewEncoder(w).Encode(map[string]any{}); err != nil {
		writeError(w, "Ошибка при отправке ответа")
	}
}

// HandleTaskDone завершает задачу
func (h *Handler) HandleTaskDone(w http.ResponseWriter, r *http.Request) {
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

	// Получаем задачу из базы данных
	task, err := db.GetTaskByID(h.DB, taskID)
	if err != nil {
		writeError(w, "Ошибка при получении задачи")
		return
	}

	if task.Repeat == "" {
		// Если задача одноразовая, удаляем её
		_, err = db.DeleteTask(h.DB, taskID)
		if err != nil {
			writeError(w, "Не удалось удалить задачу")
			return
		}
	} else {
		// Если задача повторяющаяся, обновляем дату
		now := utils.NormalizeDate(time.Now())
		nextDate, err := utils.NextDate(now, task.Date, task.Repeat)
		if err != nil {
			writeError(w, "Ошибка при расчёте следующей даты")
			return
		}

		task.Date = nextDate
		_, err = db.UpdateTask(h.DB, *task)
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
func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request) {
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

	// Удаляем задачу из базы данных через db.DeleteTask
	rowsAffected, err := db.DeleteTask(h.DB, taskID)
	if err != nil {
		writeError(w, "Не удалось удалить задачу")
		return
	}

	// Проверяем, была ли удалена задача
	if rowsAffected == 0 {
		writeError(w, "Задача не найдена")
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
