package main

import (
	"log"
	"net/http"
	"os"

	"go_final_project/db" // Пакет для работы с базой данных
)

func main() {
	// Указываем директорию для файлов фронтенда
	webDir := "./web"

	// Создаём файловый сервер с маршрутами
	fileServer := http.FileServer(http.Dir(webDir))
	http.Handle("/", fileServer)

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
