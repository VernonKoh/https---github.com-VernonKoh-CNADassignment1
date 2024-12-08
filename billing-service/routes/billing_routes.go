package routes

import (
	"cnad_assignment/billing-service/handlers" // Ensure this import is correct

	"github.com/gorilla/mux"
)

// RegisterBillingRoutes registers routes related to billing
func RegisterBillingRoutes(router *mux.Router) {
	// Register the FetchBookings route for fetching all bookings for a user
	router.HandleFunc("/api/v1/bookings", handlers.FetchBookings).Methods("GET")

	// Register the FetchBillingDetails route for fetching billing details
	router.HandleFunc("/api/v1/billing", handlers.FetchBillingDetails).Methods("GET")

	// Make sure the /api/v1/billing/bookings is handled correctly, assuming you want separate functionality
	router.HandleFunc("/api/v1/billing/bookings", handlers.FetchBillingDetails).Methods("GET") // <- Updated to match billing details
}
