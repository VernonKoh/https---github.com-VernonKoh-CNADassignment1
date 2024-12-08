package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

var DB *sql.DB

func InitDB() {
	var err error

	// Connect to MySQL
	DB, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/car_sharing")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Verify the connection
	if err = DB.Ping(); err != nil {
		log.Fatalf("Database connection error: %v", err)
	}

	log.Println("Database connected successfully")
}
