package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"go_final_project/db"    // Пакет для работы с базой данных
	"go_final_project/utils" // Пакет с функцией NextDate
)

func main() {
	// Указываем директорию для файлов фронтенда
	webDir := "./web"

	// Создаём файловый сервер с маршрутами
	fileServer := http.FileServer(http.Dir(webDir))
	http.Handle("/", fileServer)

	// Регистрируем обработчик API /api/nextdate
	http.HandleFunc("/api/nextdate", handleNextDate)

	// Получаем порт из переменной окружения (со звёздочкой) или используем порт по умолчанию
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540" // Порт по умолчанию
	}

	// Проверяем базу данных и создаём её при необходимости
	dbFile := db.GetDatabasePath() // Путь к базе данных
	err := db.SetupDatabase(dbFile)
	if err != nil {
		log.Fatalf("Error with database: %v", err)
	}

	// Запускаем сервер
	log.Printf("Starting server on :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
}

// handleNextDate обрабатывает запросы к API /api/nextdate
func handleNextDate(w http.ResponseWriter, r *http.Request) {
	// Получаем параметры запроса
	nowStr := r.FormValue("now")
	dateStr := r.FormValue("date")
	repeat := r.FormValue("repeat")

	// Парсим текущую дату
	now, err := time.Parse("20060102", nowStr)
	if err != nil {
		http.Error(w, "invalid now parameter", http.StatusBadRequest)
		return
	}

	// Вызываем функцию NextDate
	nextDate, err := utils.NextDate(now, dateStr, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Возвращаем следующую дату как чистую строку
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(nextDate))
}
