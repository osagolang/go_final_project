package handlers

import (
	"net/http"
	"time"

	"go_final_project/utils"
)

// HandleDate обрабатывает GET-запрос для следующей даты
func HandleDate(w http.ResponseWriter, r *http.Request) {
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "Invalid now parameter", http.StatusBadRequest)
		return
	}

	nextDate, err := utils.NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Упрощенный ответ для прохождения тестов
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDate))
}
