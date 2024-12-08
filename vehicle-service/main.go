package main

import (
	"cnad_assignment/vehicle-service/database"
	"cnad_assignment/vehicle-service/routes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// Initialize the database
	database.InitDB()

	// Create a new router
	router := mux.NewRouter()

	// Register vehicle routes
	routes.RegisterVehicleRoutes(router)

	// Add CORS middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8081"}, // Allow requests from this origin
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	// Wrap the router with CORS middleware
	handler := c.Handler(router)

	// Start the real-time availability checker in a separate Goroutine
	go startAvailabilityChecker()

	// Start the server
	port := ":8082" // Use a different port to avoid conflicts with the user-service
	fmt.Printf("Vehicle service is running on http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, handler))
}

// startAvailabilityChecker periodically checks and updates vehicle availability
func startAvailabilityChecker() {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Checking and updating vehicle availability...")
			if err := database.CheckAndUpdateAvailability(); err != nil {
				log.Printf("Error updating vehicle availability: %v", err)
			} else {
				log.Println("Vehicle availability updated successfully.")
			}
		}
	}
}
