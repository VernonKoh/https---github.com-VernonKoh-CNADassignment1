package routes

import (
	"cnad_assignment/vehicle-service/handlers"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func RegisterVehicleRoutes(router *mux.Router) {
	// Enable CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:8081"}, // Allow frontend on localhost:8081
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	})

	// Wrap your routes with the CORS middleware
	vehicleRouter := router.PathPrefix("/api/v1").Subrouter()
	vehicleRouter.HandleFunc("/vehicles", handlers.GetAvailableVehicles).Methods("GET")
	vehicleRouter.HandleFunc("/vehicles/{id:[0-9]+}/book", handlers.BookVehicle).Methods("POST")
	vehicleRouter.HandleFunc("/vehicles/{id:[0-9]+}/status", handlers.GetVehicleStatus).Methods("GET")
	vehicleRouter.HandleFunc("/vehicles/{id:[0-9]+}/bookings", handlers.GetBookingsForVehicle).Methods("GET")
	vehicleRouter.HandleFunc("/bookings", handlers.GetBookings).Methods("GET")
	vehicleRouter.HandleFunc("/bookings/{id}", handlers.ModifyBooking).Methods("PUT")
	vehicleRouter.HandleFunc("/bookings/{id}", handlers.CancelBooking).Methods("DELETE")

	// Register the route for fetching rental history by user ID
	router.HandleFunc("/api/v1/users/{id}/rental-history", handlers.FetchRentalHistoryByUser).Methods("GET")
	// Apply CORS middleware
	c.Handler(vehicleRouter)
}
