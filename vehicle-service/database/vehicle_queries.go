package database

import (
	"cnad_assignment/vehicle-service/models"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

func FetchAvailableVehicles() ([]models.Vehicle, error) {
	query := "SELECT id, make, model, registration_number, is_available FROM vehicles WHERE is_available = TRUE"
	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vehicles []models.Vehicle
	for rows.Next() {
		var v models.Vehicle
		if err := rows.Scan(&v.ID, &v.Make, &v.Model, &v.RegistrationNumber, &v.IsAvailable); err != nil {
			return nil, err
		}
		vehicles = append(vehicles, v)
	}
	return vehicles, nil
}

func CreateBooking(vehicleID int, booking models.Booking) error {
	tx, err := DB.Begin() // Begin a database transaction
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}

	// Check for overlapping bookings
	query := `
        SELECT id, start_time, end_time 
        FROM bookings 
        WHERE vehicle_id = ? 
          AND status = 'confirmed' 
          AND (
            (start_time < ? AND end_time > ?) OR
            (start_time < ? AND end_time > ?) OR
            (start_time >= ? AND end_time <= ?)
          )
    `
	var conflictID int
	var conflictStartTime, conflictEndTime time.Time
	err = tx.QueryRow(query, vehicleID, booking.EndTime, booking.EndTime, booking.StartTime, booking.StartTime, booking.StartTime, booking.EndTime).
		Scan(&conflictID, &conflictStartTime, &conflictEndTime)

	if err != nil && err != sql.ErrNoRows {
		tx.Rollback()
		log.Printf("Error checking overlapping bookings: %v", err)
		return fmt.Errorf("failed to check overlapping bookings: %v", err)
	}

	if err == nil { // Conflict found
		tx.Rollback()
		log.Printf("Booking conflict: overlapping time range for vehicle ID=%d", vehicleID)
		return fmt.Errorf("time range overlaps with an existing booking from %v to %v", conflictStartTime, conflictEndTime)
	}

	// Insert the booking into the database
	insertQuery := "INSERT INTO bookings (user_id, vehicle_id, start_time, end_time, status) VALUES (?, ?, ?, ?, ?)"
	_, err = tx.Exec(insertQuery, booking.UserID, vehicleID, booking.StartTime, booking.EndTime, "confirmed")
	if err != nil {
		tx.Rollback()
		log.Printf("Error inserting booking: %v", err)
		return fmt.Errorf("failed to insert booking: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	log.Printf("Booking created successfully for vehicle ID=%d and user ID=%d", vehicleID, booking.UserID)
	return nil
}

func FetchVehicleStatus(vehicleID int) (models.VehicleStatus, error) {
	var status models.VehicleStatus
	query := "SELECT vehicle_id, location, charge_level, cleanliness, updated_at FROM vehicle_status WHERE vehicle_id = ?"
	err := DB.QueryRow(query, vehicleID).Scan(&status.VehicleID, &status.Location, &status.ChargeLevel, &status.Cleanliness, &status.UpdatedAt)
	if err == sql.ErrNoRows {
		return status, errors.New("vehicle status not found")
	}
	return status, err
}

// CheckAndUpdateAvailability updates bookings and vehicle availability for expired bookings
func CheckAndUpdateAvailability() error {
	now := time.Now()
	log.Printf("Checking expired bookings at: %v", now)

	// Query to find expired bookings with their `end_time`
	query := `
		SELECT id, vehicle_id, end_time
		FROM bookings
		WHERE end_time < ? AND status = 'confirmed'
	`
	rows, err := DB.Query(query, now)
	if err != nil {
		log.Printf("Error querying expired bookings: %v", err)
		return err
	}
	defer rows.Close()

	var updates []struct {
		BookingID int
		VehicleID int
		EndTime   time.Time
	}

	// Collect all expired bookings
	for rows.Next() {
		var bookingID, vehicleID int
		var endTime time.Time
		if err := rows.Scan(&bookingID, &vehicleID, &endTime); err != nil {
			log.Printf("Error scanning booking row: %v", err)
			return err
		}

		// Log each expired booking for debugging
		log.Printf("Booking ID: %d, Vehicle ID: %d, End Time: %v, Now: %v", bookingID, vehicleID, endTime, now)
		updates = append(updates, struct {
			BookingID int
			VehicleID int
			EndTime   time.Time
		}{BookingID: bookingID, VehicleID: vehicleID, EndTime: endTime})
	}

	if len(updates) == 0 {
		log.Println("No expired bookings found.")
		return nil
	}

	log.Printf("Expired bookings found: %+v", updates)

	// Update vehicle availability and mark bookings as "completed"
	tx, err := DB.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}

	for _, update := range updates {
		// Mark the booking as 'completed'
		updateBookingQuery := "UPDATE bookings SET status = 'completed' WHERE id = ?"
		_, err = tx.Exec(updateBookingQuery, update.BookingID)
		if err != nil {
			log.Printf("Error updating booking status for booking ID %d: %v", update.BookingID, err)
			tx.Rollback()
			return err
		}

		// Mark the vehicle as 'available'
		updateVehicleQuery := "UPDATE vehicles SET is_available = TRUE WHERE id = ?"
		_, err = tx.Exec(updateVehicleQuery, update.VehicleID)
		if err != nil {
			log.Printf("Error updating vehicle availability for vehicle ID %d: %v", update.VehicleID, err)
			tx.Rollback()
			return err
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Error committing transaction: %v", err)
		return err
	}

	log.Println("Vehicle availability updated successfully.")
	return nil
}

func FetchBookingsForVehicle(vehicleID int) ([]models.Booking, error) {
	query := "SELECT id, user_id, vehicle_id, start_time, end_time, status FROM bookings WHERE vehicle_id = ? AND status = 'confirmed'"
	rows, err := DB.Query(query, vehicleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []models.Booking
	for rows.Next() {
		var booking models.Booking
		err := rows.Scan(&booking.ID, &booking.UserID, &booking.VehicleID, &booking.StartTime, &booking.EndTime, &booking.Status)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}
	return bookings, nil
}

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

		err := rows.Scan(&bookingID, &userID, &startTime, &endTime, &status, &make, &model, &registrationNumber)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}

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

	if len(bookings) == 0 {
		log.Printf("No bookings found for user %d", userID)
	}

	return bookings, nil
}

func ModifyBooking(bookingID int, newStartTime, newEndTime time.Time) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Check for overlapping bookings
	checkQuery := `
        SELECT COUNT(*) 
        FROM bookings 
        WHERE vehicle_id = (SELECT vehicle_id FROM bookings WHERE id = ?) 
          AND id != ?
          AND status IN ('confirmed', 'modified')
          AND (
            (start_time < ? AND end_time > ?) OR
            (start_time < ? AND end_time > ?) OR
            (start_time >= ? AND end_time <= ?)
          )
    `
	var overlapCount int
	err = tx.QueryRow(checkQuery, bookingID, bookingID, newEndTime, newEndTime, newStartTime, newStartTime, newStartTime, newEndTime).Scan(&overlapCount)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to check overlapping bookings: %v", err)
	}

	if overlapCount > 0 {
		tx.Rollback()
		return fmt.Errorf("overlapping booking exists")
	}

	// Update the booking
	updateQuery := `
        UPDATE bookings
        SET start_time = ?, end_time = ?, status = 'modified'
        WHERE id = ?
    `
	_, err = tx.Exec(updateQuery, newStartTime, newEndTime, bookingID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update booking: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func CancelBooking(bookingID int) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	// Update the booking status
	updateQuery := `
        UPDATE bookings
        SET status = 'canceled'
        WHERE id = ?
    `
	_, err = tx.Exec(updateQuery, bookingID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func FetchRentalHistoryByUser(userID int) ([]map[string]interface{}, error) {
	// Query for all bookings, including past ones
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
        WHERE b.user_id = ?
    `

	// Execute the query to fetch all bookings
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

		// Scan the result into the variables
		err := rows.Scan(&bookingID, &userID, &startTime, &endTime, &status, &make, &model, &registrationNumber)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			return nil, err
		}

		// Append the booking information into the bookings slice
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

	if len(bookings) == 0 {
		log.Printf("No bookings found for user %d", userID)
	}

	return bookings, nil
}
