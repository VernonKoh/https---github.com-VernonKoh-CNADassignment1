package handlers

import (
	"cnad_assignment/billing-service/database"
	"cnad_assignment/billing-service/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Define the User struct to match the user API response structure
type User struct {
	ID   int    `json:"id"`
	Role string `json:"role"`
}

// FetchBillingDetails fetches the billing details for a user including booking and vehicle details
func FetchBillingDetails(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Fetch the bookings and vehicle details for the user
	bookings, err := database.FetchBookingsByUser(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching bookings: %v", err), http.StatusInternalServerError)
		return
	}

	// Calculate total cost and gather billing details
	var totalCost float64
	var billingDetails []map[string]interface{}

	// Process each booking
	for _, booking := range bookings {
		// Get the vehicle details
		vehicle := fmt.Sprintf("%s %s (%s)", booking["make"], booking["model"], booking["registration_number"])

		// Calculate the cost for the booking
		cost, err := utils.CalculateBilling(userID, booking["start_time"].(time.Time), booking["end_time"].(time.Time))
		if err != nil {
			http.Error(w, fmt.Sprintf("Error calculating billing: %v", err), http.StatusInternalServerError)
			return
		}

		// Add the billing details for each booking to the response
		billingDetails = append(billingDetails, map[string]interface{}{
			"vehicle":              vehicle,
			"start_time":           booking["start_time"],
			"end_time":             booking["end_time"],
			"duration":             booking["end_time"].(time.Time).Sub(booking["start_time"].(time.Time)).Hours(),
			"cost_before_discount": fmt.Sprintf("$%.2f", cost),
			"discount":             booking["discount"],
			"final_cost":           fmt.Sprintf("$%.2f", cost),
		})

		// Add the cost to the total cost
		totalCost += cost
	}

	// Respond with the total cost and billing details
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_cost":      totalCost,
		"billing_details": billingDetails,
	})
}

// FetchBookings fetches all bookings for a specific user
func FetchBookings(w http.ResponseWriter, r *http.Request) {
	// Get the user ID from URL query parameters
	userIDStr := r.URL.Query().Get("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Fetch the bookings from the vehicle-service API
	resp, err := http.Get(fmt.Sprintf("http://localhost:8082/api/v1/bookings?user_id=%d", userID)) // Adjust URL as per vehicle service
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching bookings: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Parse the bookings response
	var bookings []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&bookings); err != nil {
		http.Error(w, "Failed to decode bookings", http.StatusInternalServerError)
		return
	}

	// Check if bookings exist and handle them properly
	if len(bookings) == 0 {
		http.Error(w, "No bookings found", http.StatusNotFound)
		return
	}

	// Calculate the total cost based on bookings
	var totalCost float64
	for _, booking := range bookings {
		startTime, _ := time.Parse(time.RFC3339, booking["start_time"].(string))
		endTime, _ := time.Parse(time.RFC3339, booking["end_time"].(string))

		// Calculate billing for each booking
		cost, err := utils.CalculateBilling(userID, startTime, endTime)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error calculating billing: %v", err), http.StatusInternalServerError)
			return
		}
		totalCost += cost
	}

	// Send back the total cost and bookings info
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_cost": totalCost,
		"bookings":   bookings,
	})
}
