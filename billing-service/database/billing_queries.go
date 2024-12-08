package database

import (
	"cnad_assignment/vehicle-service/models"
	"database/sql"
	"fmt"
	"log"
	"time"
)

func FetchBookingsByUser(userID int) ([]map[string]interface{}, error) {
	query := `
        SELECT 
            b.id AS booking_id, 
            b.user_id, 
            b.start_time, 
            b.end_time, 
            b.status, 
            v.make, 
            v.model, 
            v.registration_number 
        FROM bookings b 
        JOIN vehicles v ON b.vehicle_id = v.id 
        WHERE b.user_id = ? AND b.status IN ('confirmed', 'modified');
    `
	rows, err := DB.Query(query, userID)
	if err != nil {
		log.Printf("Error executing query for user %d: %v", userID, err)
		return nil, err
	}
	defer rows.Close()

	var bookings []map[string]interface{}
	for rows.Next() {
		var bookingID, userID int
		var startTime, endTime, status, make, model, registrationNumber string

		// Scan the result into variables
		err := rows.Scan(&bookingID, &userID, &startTime, &endTime, &status, &make, &model, &registrationNumber)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}

		// Use time.RFC3339 to parse times with timezone offsets
		startTimeParsed, err := time.Parse(time.RFC3339, startTime)
		if err != nil {
			log.Printf("Error parsing start time: %v", err)
			return nil, err
		}

		endTimeParsed, err := time.Parse(time.RFC3339, endTime)
		if err != nil {
			log.Printf("Error parsing end time: %v", err)
			return nil, err
		}

		// Append the booking details along with vehicle details
		bookings = append(bookings, map[string]interface{}{
			"booking_id":          bookingID,
			"user_id":             userID,
			"start_time":          startTimeParsed,
			"end_time":            endTimeParsed,
			"status":              status,
			"make":                make,
			"model":               model,
			"registration_number": registrationNumber,
		})
	}

	if len(bookings) == 0 {
		log.Printf("No bookings found for user %d", userID)
	}

	return bookings, nil
}

// FetchVehicleDetails fetches the vehicle details based on vehicleID
func FetchVehicleDetails(vehicleID int) (models.Vehicle, error) {
	var vehicle models.Vehicle
	query := "SELECT make, model, registration_number FROM vehicles WHERE id = ?"
	err := DB.QueryRow(query, vehicleID).Scan(&vehicle.Make, &vehicle.Model, &vehicle.RegistrationNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return vehicle, fmt.Errorf("no vehicle found with id %d", vehicleID)
		}
		log.Printf("Error fetching vehicle details: %v", err)
		return vehicle, err
	}
	return vehicle, nil
}
