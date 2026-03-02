package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Читаем файл миграции
	content, err := os.ReadFile("migrations/create_table.up.sql")
	if err != nil {
		log.Fatal("Failed to read migration file:", err)
	}

	// Подключаемся к базе
	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/subscription_db?sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Выполняем SQL
	if _, err := db.Exec(string(content)); err != nil {
		log.Fatal("Failed to execute migration:", err)
	}

	log.Println("✅ Migration completed successfully!")
}
