package handlers

import "database/sql"

// Handler - структура для хранения зависимостей обработчиков
type Handler struct {
	DB *sql.DB
}

// NewHandler создаёт новый экземпляр Handler
func NewHandler(db *sql.DB) *Handler {
	return &Handler{DB: db}
}
