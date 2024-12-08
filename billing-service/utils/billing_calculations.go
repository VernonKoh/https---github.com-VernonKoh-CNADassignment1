package utils

import (
	"cnad_assignment/billing-service/database"
	"fmt"
	"math"
	"time"
)

// CalculateBilling calculates the cost based on membership level and rental duration
func CalculateBilling(userID int, startTime, endTime time.Time) (float64, error) {
	var hourlyRateDiscount float64
	var role string

	// Fetch the user's role (membership tier) from the database
	err := database.DB.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)
	if err != nil {
		return 0, err
	}

	// Set discount based on membership tier
	switch role {
	case "Premium":
		hourlyRateDiscount = 0.10 // 10% discount for Premium
	case "VIP":
		hourlyRateDiscount = 0.20 // 20% discount for VIP
	default:
		hourlyRateDiscount = 0.00 // No discount for Basic tier
	}

	// Calculate rental duration in hours
	rentalDuration := endTime.Sub(startTime).Hours()

	// If rental duration is negative, it's invalid
	if rentalDuration <= 0 {
		return 0, fmt.Errorf("invalid rental duration")
	}

	// Base rate per hour
	baseRate := 20.00 // Example: $20 per hour

	// Apply discount
	discountedRate := baseRate * (1 - hourlyRateDiscount)

	// Calculate the total cost for the rental
	totalCost := discountedRate * rentalDuration

	// Round to 2 decimal places for proper currency formatting
	return math.Round(totalCost*100) / 100, nil
}
