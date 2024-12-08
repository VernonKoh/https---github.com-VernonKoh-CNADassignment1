package main

import (
	"cnad_assignment/billing-service/database"
	"cnad_assignment/billing-service/routes"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	// Initialize the database
	database.InitDB()

	// Create a new router
	router := mux.NewRouter()

	// Register API routes for billing-related functionality
	routes.RegisterBillingRoutes(router)

	// Enable CORS for all origins
	// You can also restrict this to specific domains by modifying the AllowedOrigins list
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:8081"}), // Specify the origin for your frontend
		handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)(router)

	// Start the server
	fmt.Println("Billing service is running on http://localhost:8083")
	log.Fatal(http.ListenAndServe(":8083", corsHandler))
}
