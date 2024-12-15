package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	// Указываем директорию для файлов фронтенда
	webDir := "./web"

	// Создаём файловый сервер
	fileServer := http.FileServer(http.Dir(webDir))

	// Регистрируем маршруты
	http.Handle("/", fileServer)

	// Получаем порт из переменной окружения или используем порт по умолчанию
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
