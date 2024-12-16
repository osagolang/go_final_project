package main

import (
	"log"
	"net/http"
	"os"

	"go_final_project/db"
	"go_final_project/handlers"
)

func main() {
	// Указываем директорию для файлов фронтенда
	webDir := "./web"
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	// Устанавливаем маршруты
	http.HandleFunc("/api/task", handlers.HandleTask)      // Для добавления задачи
	http.HandleFunc("/api/nextdate", handlers.HandleDate)  // Для расчёта следующей даты
	http.HandleFunc("/api/tasks", handlers.HandleTaskList) // Для списка задач

	// Получаем порт из переменной окружения (Задача со звёздочкой)
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540" // Порт по умолчанию
	}

	// Проверяем и создаём базу данных при необходимости
	dbFile := db.GetDatabasePath()
	if err := db.SetupDatabase(dbFile); err != nil {
		log.Fatalf("Error with database: %v", err)
	}

	// Запускаем сервер
	log.Printf("Starting server on :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
}
