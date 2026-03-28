package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func Connect() *sql.DB {
	// ВНИМАНИЕ: Проверь логин/пароль к своей локальной БД PostgreSQL
	connStr := "postgres://postgres:password@localhost:5432/postgres?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Ошибка подключения:", err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal("База недоступна:", err)
	}
	return db
}
