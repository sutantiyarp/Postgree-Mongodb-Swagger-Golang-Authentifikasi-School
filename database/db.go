package database

import (
	"database/sql"
	"log"
	"os"
	_ "github.com/lib/pq" // PostgreSQL driver
)

var DB *sql.DB

func ConnectDB() *sql.DB {
	dsn := os.Getenv("DB_DSN") // Ambil nilai dari .env
	if dsn == "" {
		log.Fatal("DB_DSN environment variable is missing")
	}

	// Jangan tambahkan sslmode=disable lagi di sini!
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}
	log.Println("Connected to database successfully")
	return db
}
