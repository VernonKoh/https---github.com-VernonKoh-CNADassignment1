package database

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql" // Import MySQL driver to establish database connection
)

// DB is a global variable that holds the database connection pool.
// It is accessible throughout the application and is used to interact with the MySQL database.
var DB *sql.DB

// InitDB initializes the database connection.
// This function establishes a connection to the MySQL database and verifies the connection.
// It is called during the application's startup to ensure that the database is available for use.
func InitDB() {
	var err error

	// Attempt to open a connection to the MySQL database.
	// The connection string format is "username:password@tcp(host:port)/database_name".
	// Replace "user:password" with actual credentials, "127.0.0.1:3306" with the MySQL server's address and port,
	// and "car_sharing" with the actual database name.
	DB, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/car_sharing")
	if err != nil {
		// If an error occurs while opening the connection, log the error and terminate the application.
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Verify the connection by attempting to ping the database.
	// If the database is unreachable or the connection is invalid, this will return an error.
	if err = DB.Ping(); err != nil {
		// If an error occurs during the ping operation, log the error and terminate the application.
		log.Fatalf("Database connection error: %v", err)
	}

	// If the connection is successful, log a message indicating that the database has been connected.
	log.Println("Database connected successfully")
}
