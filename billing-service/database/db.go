package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql" // Import MySQL driver
)

var DB *sql.DB

// InitDB initializes the database connection for the billing service
func InitDB() {
	var err error

	// Replace "user:password@tcp(127.0.0.1:3306)/car_sharing" with your database credentials
	DB, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/car_sharing?parseTime=true&loc=Local")
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Verify the connection
	if err = DB.Ping(); err != nil {
		log.Fatalf("Database connection error: %v", err)
	}

	log.Println("Database connected successfully")
}
