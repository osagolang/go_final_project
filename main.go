package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"go_final_project/db"
	"go_final_project/handlers"
)

func main() {
	// Указываем директорию для файлов фронтенда
	webDir := "./web"
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	// Инициализация пути к базе данных
	dbPath := os.Getenv("TODO_DBFILE")
	if dbPath == "" {
		workingDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Не удалось получить рабочую директорию: %v", err)
		}
		dbPath = filepath.Join(workingDir, "scheduler.db")
	}

	// Проверяем и создаём базу данных при необходимости
	if err := db.SetupDatabase(dbPath); err != nil {
		log.Fatalf("Error with database: %v", err)
	}

	// Инициализация подключения к базе данных
	dbConn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("Не удалось подключиться к базе данных: %v", err)
	}
	defer dbConn.Close() // Закрываем подключение при завершении программы

	// Инициализируем обработчики с передачей подключения к базе данных
	handler := handlers.NewHandler(dbConn)

	// Устанавливаем маршруты
	http.HandleFunc("/api/task", handler.HandleTask)          // Для действий с задачами
	http.HandleFunc("/api/nextdate", handlers.HandleDate)     // Для расчёта следующей даты
	http.HandleFunc("/api/tasks", handler.HandleTaskList)     // Для списка задач
	http.HandleFunc("/api/task/done", handler.HandleTaskDone) // Для завершения задачи

	// Получаем порт из переменной окружения (Задача со звёздочкой)
	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540" // Порт по умолчанию
	}

	// Запускаем сервер
	log.Printf("Starting server on :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Error starting server: %v\n", err)
	}
}
