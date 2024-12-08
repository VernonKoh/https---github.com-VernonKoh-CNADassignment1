package routes

import (
	"cnad_assignment/user-service/handlers"

	"github.com/gorilla/mux"
)

// RegisterUserRoutes sets up the API routes for User Management
func RegisterUserRoutes(router *mux.Router) {
	userRouter := router.PathPrefix("/api/v1/users").Subrouter()
	userRouter.HandleFunc("/register", handlers.RegisterUser).Methods("POST")
	userRouter.HandleFunc("/login", handlers.LoginUser).Methods("POST")
	userRouter.HandleFunc("/verify", handlers.VerifyUser).Methods("GET") // Add this route for verification
	userRouter.HandleFunc("/membership-tiers", handlers.GetMembershipTiers).Methods("GET")

	userRouter.HandleFunc("/{id}", handlers.GetUserProfile).Methods("GET")
	userRouter.HandleFunc("/{id}", handlers.UpdateUserProfile).Methods("PUT")
	userRouter.HandleFunc("/{id}/membership-benefits", handlers.GetUserMembershipBenefits).Methods("GET")
}
