package handlers

import (
	"cnad_assignment/vehicle-service/database"
	"cnad_assignment/vehicle-service/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func GetAvailableVehicles(w http.ResponseWriter, r *http.Request) {
	log.Println("GetAvailableVehicles called") // Add this log for debugging
	vehicles, err := database.FetchAvailableVehicles()
	if err != nil {
		log.Printf("Error fetching available vehicles: %v", err) // Log the error
		http.Error(w, "Failed to fetch vehicles", http.StatusInternalServerError)
		return
	}
	log.Printf("Fetched vehicles: %+v", vehicles) // Log the vehicles fetched
	json.NewEncoder(w).Encode(vehicles)
}

func BookVehicle(w http.ResponseWriter, r *http.Request) {
	log.Println("BookVehicle handler called")

	vehicleID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		log.Printf("Invalid vehicle ID: %v", err)
		http.Error(w, "Invalid vehicle ID", http.StatusBadRequest)
		return
	}

	var bookingRequest struct {
		UserID    int    `json:"user_id"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
	}

	if err := json.NewDecoder(r.Body).Decode(&bookingRequest); err != nil {
		log.Printf("Invalid input data: %v", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	log.Printf("Booking request: %+v", bookingRequest)

	loc, _ := time.LoadLocation("Local") // Ensure local timezone

	// Parse incoming UTC time and convert to local time
	startTimeUTC, err := time.Parse(time.RFC3339, bookingRequest.StartTime)
	if err != nil {
		log.Printf("Invalid start time format: %v", err)
		http.Error(w, "Invalid start time format", http.StatusBadRequest)
		return
	}
	endTimeUTC, err := time.Parse(time.RFC3339, bookingRequest.EndTime)
	if err != nil {
		log.Printf("Invalid end time format: %v", err)
		http.Error(w, "Invalid end time format", http.StatusBadRequest)
		return
	}

	// Convert UTC times to local times
	startTime := startTimeUTC.In(loc)
	endTime := endTimeUTC.In(loc)
	now := time.Now().In(loc)

	// Validate that start time is not in the past and end time is after start time
	if startTime.Before(now) {
		log.Printf("Start time %v is in the past. Current time: %v", startTime, now)
		http.Error(w, "Start time cannot be in the past", http.StatusBadRequest)
		return
	}
	if startTime.After(endTime) {
		log.Printf("Start time %v is after end time %v", startTime, endTime)
		http.Error(w, "End time must be after start time", http.StatusBadRequest)
		return
	}

	booking := models.Booking{
		UserID:    bookingRequest.UserID,
		VehicleID: vehicleID,
		StartTime: startTime,
		EndTime:   endTime,
		Status:    "confirmed",
	}

	log.Printf("Attempting to book vehicle ID=%d for user ID=%d", vehicleID, bookingRequest.UserID)

	if err := database.CreateBooking(vehicleID, booking); err != nil {
		log.Printf("Error creating booking: %v", err)
		if strings.Contains(err.Error(), "time range overlaps") {
			// Send a structured JSON response for the conflict
			conflictDetails := strings.Split(err.Error(), " from ")
			conflictStartEnd := strings.Split(conflictDetails[1], " to ")
			http.Error(w, fmt.Sprintf(`{"error": "Time range overlaps with an existing booking", "conflict_start_time": "%s", "conflict_end_time": "%s"}`, conflictStartEnd[0], conflictStartEnd[1]), http.StatusConflict)
		} else {
			http.Error(w, "Failed to book vehicle", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("Vehicle %d successfully booked by user %d", vehicleID, bookingRequest.UserID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Vehicle booked successfully"})
}

func GetVehicleStatus(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid vehicle ID", http.StatusBadRequest)
		return
	}

	status, err := database.FetchVehicleStatus(vehicleID)
	if err != nil {
		http.Error(w, "Failed to fetch vehicle status", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(status)
}
func GetBookingsForVehicle(w http.ResponseWriter, r *http.Request) {
	vehicleID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "Invalid vehicle ID", http.StatusBadRequest)
		return
	}

	bookings, err := database.FetchBookingsForVehicle(vehicleID)
	if err != nil {
		http.Error(w, "Failed to fetch bookings", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(bookings)
}

func GetBookings(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.Atoi(r.URL.Query().Get("user_id"))
	if err != nil || userID <= 0 {
		log.Printf("Invalid user ID: %v", err)
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	log.Printf("Fetching bookings for user ID: %d", userID)

	bookings, err := database.FetchBookingsByUser(userID)
	if err != nil {
		log.Printf("Error fetching bookings for user %d: %v", userID, err)
		http.Error(w, "Failed to fetch bookings", http.StatusInternalServerError)
		return
	}

	if len(bookings) == 0 {
		log.Printf("No bookings found for user %d", userID)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "No bookings found."})
		return
	}

	json.NewEncoder(w).Encode(bookings)
}

func ModifyBooking(w http.ResponseWriter, r *http.Request) {
	bookingID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || bookingID <= 0 {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	var updateRequest struct {
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
	}
	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	loc, _ := time.LoadLocation("Local")
	startTime, err := time.ParseInLocation("2006-01-02T15:04:05", updateRequest.StartTime, loc)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid start time format"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	endTime, err := time.ParseInLocation("2006-01-02T15:04:05", updateRequest.EndTime, loc)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid end time format"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now().In(loc)

	// Validate start time and end time
	if startTime.Before(now) {
		json.NewEncoder(w).Encode(map[string]string{"error": "New start time cannot be earlier than the current time"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if startTime.After(endTime) {
		json.NewEncoder(w).Encode(map[string]string{"error": "End time cannot be earlier than the start time"})
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := database.ModifyBooking(bookingID, startTime, endTime); err != nil {
		log.Printf("Error modifying booking: %v", err)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to modify booking due to server error"})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Booking modified successfully"})
}

func CancelBooking(w http.ResponseWriter, r *http.Request) {
	bookingID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || bookingID <= 0 {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	// Cancel booking in the database
	if err := database.CancelBooking(bookingID); err != nil {
		log.Printf("Error canceling booking: %v", err)
		http.Error(w, "Failed to cancel booking", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Booking canceled successfully"})
}

// FetchRentalHistoryByUser fetches all past and present bookings of a user
func FetchRentalHistoryByUser(w http.ResponseWriter, r *http.Request) {
	// Get the userID from the URL params
	userID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil || userID <= 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Fetch the user's rental history from the database
	bookings, err := database.FetchRentalHistoryByUser(userID)
	if err != nil {
		log.Printf("Error fetching rental history for user %d: %v", userID, err)
		http.Error(w, "Failed to fetch rental history", http.StatusInternalServerError)
		return
	}

	// Encode and return the response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(bookings)
}
