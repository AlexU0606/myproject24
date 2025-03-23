package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

// Глобальная переменная для БД
var DB *sql.DB

// InitDB инициализирует соединение с базой данных
func InitDB() {
	var err error
	DB, err = sql.Open("sqlite", "./medicine.db")
	if err != nil {
		log.Fatal("Ошибка подключения к БД:", err)
	}

	// Создание таблицы, если её нет
	createTable()
}

// createTable создает таблицу в БД
func createTable() {
	query := `CREATE TABLE IF NOT EXISTS schedules (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT,
		medicine TEXT,
		frequency INTEGER,
		duration INTEGER,
		created_at TEXT,
		times  TEXT
	)`
	_, err := DB.Exec(query)
	if err != nil {
		log.Fatal("Ошибка создания таблицы:", err)
	}
}

// CloseDB закрывает соединение с БД
func CloseDB() {
	DB.Close()
}
