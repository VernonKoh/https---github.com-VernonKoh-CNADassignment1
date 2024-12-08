package database

import (
	"log"
	"time"
)

// FetchBookingsByUser fetches bookings for a specific user by their userID.
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
        WHERE b.user_id = ? AND b.status IN ('confirmed', 'modified', 'completed');` // Make sure 'completed' is considered too

	rows, err := DB.Query(query, userID)
	if err != nil {
		log.Printf("Error executing query for user %d: %v", userID, err)
		return nil, err
	}
	defer rows.Close()

	var bookings []map[string]interface{}
	for rows.Next() {
		var bookingID, userID int
		var startTimeStr, endTimeStr, status, make, model, registrationNumber string

		// Scan the result into variables
		err := rows.Scan(&bookingID, &userID, &startTimeStr, &endTimeStr, &status, &make, &model, &registrationNumber)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}

		// Parse the start_time and end_time from string to time.Time
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			log.Printf("Error parsing start time: %v", err)
			return nil, err
		}

		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			log.Printf("Error parsing end time: %v", err)
			return nil, err
		}

		// Append the booking details along with vehicle details to the bookings slice
		bookings = append(bookings, map[string]interface{}{
			"booking_id":          bookingID,
			"user_id":             userID,
			"start_time":          startTime,
			"end_time":            endTime,
			"status":              status,
			"make":                make,
			"model":               model,
			"registration_number": registrationNumber,
		})
	}

	// If no bookings are found for the user, log the event
	if len(bookings) == 0 {
		log.Printf("No bookings found for user %d", userID)
	}

	return bookings, nil
}
