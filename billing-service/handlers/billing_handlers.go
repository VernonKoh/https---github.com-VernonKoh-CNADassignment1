package handlers

import (
	"cnad_assignment/billing-service/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Define the User struct to match the user API response structure
type User struct {
	ID   int    `json:"id"`
	Role string `json:"role"`
}

// FetchBillingDetails fetches the booking details from the vehicle-service via HTTP
func FetchBillingDetails(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL query
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Convert user ID to integer
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Fetch bookings from the vehicle service
	resp, err := http.Get(fmt.Sprintf("http://localhost:8082/api/v1/bookings?user_id=%d", userID))
	if err != nil {
		log.Printf("Error fetching booking details: %v", err)
		http.Error(w, fmt.Sprintf("Error fetching booking details from vehicle service: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Check for non-200 responses
	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("Vehicle service returned error: %s", resp.Status), resp.StatusCode)
		return
	}

	// Parse booking data
	var bookings []struct {
		ID        int       `json:"id"`
		UserID    int       `json:"user_id"`
		VehicleID int       `json:"vehicle_id"`
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
		Status    string    `json:"status"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&bookings); err != nil {
		log.Printf("Error decoding booking details: %v", err)
		http.Error(w, "Error decoding booking details", http.StatusInternalServerError)
		return
	}

	// Fetch user details (role: Premium, VIP)
	userResp, err := http.Get(fmt.Sprintf("http://localhost:8081/api/v1/users/%d", userID))
	if err != nil {
		log.Printf("Error fetching user details: %v", err)
		http.Error(w, fmt.Sprintf("Error fetching user details: %v", err), http.StatusInternalServerError)
		return
	}
	defer userResp.Body.Close()

	// Check user response
	if userResp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("User service returned error: %s", userResp.Status), userResp.StatusCode)
		return
	}

	var userRole struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&userRole); err != nil {
		log.Printf("Error decoding user details: %v", err)
		http.Error(w, "Error decoding user details", http.StatusInternalServerError)
		return
	}

	// Process each booking and calculate costs
	var totalCost float64
	var billingDetails []map[string]interface{}

	for _, booking := range bookings {
		// Calculate the duration of the booking in hours
		duration := booking.EndTime.Sub(booking.StartTime).Hours()

		// Calculate the cost before discount
		cost, err := utils.CalculateBilling(booking.UserID, booking.StartTime, booking.EndTime)
		if err != nil {
			log.Printf("Error calculating billing for booking %d: %v", booking.ID, err)
			http.Error(w, fmt.Sprintf("Error calculating billing for booking %d: %v", booking.ID, err), http.StatusInternalServerError)
			return
		}

		// Apply discount based on membership
		discount := 0.0
		if userRole.Role == "Premium" {
			discount = 0.10 // 10% discount for Premium
		} else if userRole.Role == "VIP" {
			discount = 0.20 // 20% discount for VIP
		}

		// Calculate the final price after discount
		finalCost := cost * (1 - discount)

		// Fetch vehicle details
		vehicleResp, err := http.Get(fmt.Sprintf("http://localhost:8082/api/v1/vehicles/%d", booking.VehicleID))
		if err != nil {
			log.Printf("Error fetching vehicle details for booking %d: %v", booking.ID, err)
			http.Error(w, fmt.Sprintf("Error fetching vehicle details for booking %d: %v", booking.ID, err), http.StatusInternalServerError)
			return
		}
		defer vehicleResp.Body.Close()

		var vehicle struct {
			Make               string `json:"make"`
			Model              string `json:"model"`
			RegistrationNumber string `json:"registration_number"`
		}
		if err := json.NewDecoder(vehicleResp.Body).Decode(&vehicle); err != nil {
			log.Printf("Error decoding vehicle details for booking %d: %v", booking.ID, err)
			http.Error(w, fmt.Sprintf("Error decoding vehicle details for booking %d: %v", booking.ID, err), http.StatusInternalServerError)
			return
		}

		// Add the booking details to the response
		billingDetails = append(billingDetails, map[string]interface{}{
			"vehicle":              fmt.Sprintf("%s %s (%s)", vehicle.Make, vehicle.Model, vehicle.RegistrationNumber),
			"start_time":           booking.StartTime.Format("2006-01-02 15:04:05"),
			"end_time":             booking.EndTime.Format("2006-01-02 15:04:05"),
			"duration":             fmt.Sprintf("%.2f hours", duration),
			"cost_before_discount": fmt.Sprintf("$%.2f", cost),
			"discount":             fmt.Sprintf("%.2f%%", discount*100),
			"final_cost":           fmt.Sprintf("$%.2f", finalCost),
		})

		totalCost += finalCost
	}

	// Log the billing details for debugging purposes
	log.Printf("Billing details: %+v", billingDetails)

	// Respond with the total cost and detailed billing information
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
