package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite" // Подключаем SQLite без CGO
)

// GetDatabasePath возвращает путь к файлу базы данных.
// Если переменная TODO_DBFILE задана, используется её значение. (Задача со звёздочкой)
func GetDatabasePath() string {
	dbPath := os.Getenv("TODO_DBFILE")
	if dbPath == "" {
		workingDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current working directory: %v", err)
		}
		dbPath = filepath.Join(workingDir, "scheduler.db")
	}
	return dbPath
}

// SetupDatabase проверяет наличие файла базы данных и создаёт таблицу, если её нет.
func SetupDatabase(dbFile string) error {
	workingDir, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get working directory: %v", err)
	} else {
		log.Printf("Current working directory: %s", workingDir)
	}

	log.Printf("Database file path: %s", dbFile)

	_, err = os.Stat(dbFile)
	var install bool
	if err != nil && os.IsNotExist(err) {
		install = true
		log.Println("Database file not found, creating an empty database file.")

		file, err := os.Create(dbFile)
		if err != nil {
			return fmt.Errorf("failed to create database file: %v", err)
		}
		file.Close()
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}
	defer db.Close()

	if install {
		err = createTable(db)
		if err != nil {
			return err
		}
	}
	return nil
}

// createTable создаёт таблицу и индекс по полю date.
func createTable(db *sql.DB) error {
	log.Println("Creating table 'scheduler'...")
	query := `
	CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		title TEXT NOT NULL,
		comment TEXT,
		repeat TEXT CHECK(length(repeat) <= 128)
	);
	CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("Failed to create table: %v", err)
		return err
	}
	log.Println("Table 'scheduler' created successfully.")
	return nil
}

// AddTask добавляет новую задачу в таблицу scheduler и возвращает её ID.
func AddTask(date, title, comment, repeat string) (int64, error) {
	dbPath := GetDatabasePath()
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	query := `
		INSERT INTO scheduler (date, title, comment, repeat)
		VALUES (?, ?, ?, ?)
	`
	res, err := db.Exec(query, date, title, comment, repeat)
	if err != nil {
		log.Printf("Failed to insert task: %v", err)
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		log.Printf("Failed to retrieve last insert ID: %v", err)
		return 0, err
	}

	return id, nil
}
