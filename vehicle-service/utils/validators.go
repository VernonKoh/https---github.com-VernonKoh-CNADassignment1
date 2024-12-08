package utils

import (
	"errors"
	"time"
)

// ValidateTimeRange checks if the given start and end times are valid
func ValidateTimeRange(startTime, endTime time.Time) error {
	// Ensure the start time is before the end time
	if startTime.After(endTime) {
		return errors.New("start time must be before end time")
	}

	// Ensure the booking is not in the past
	now := time.Now()
	if startTime.Before(now) || endTime.Before(now) {
		return errors.New("booking time cannot be in the past")
	}

	return nil
}

// ValidateID ensures that the provided ID is a positive integer
func ValidateID(id int) error {
	if id <= 0 {
		return errors.New("ID must be a positive integer")
	}
	return nil
}

// ValidateChargeLevel checks if the charge level is within the accepted range
func ValidateChargeLevel(chargeLevel int) error {
	if chargeLevel < 0 || chargeLevel > 100 {
		return errors.New("charge level must be between 0 and 100")
	}
	return nil
}

// ValidateCleanliness checks if the cleanliness status is valid
func ValidateCleanliness(cleanliness string) error {
	validStatuses := map[string]bool{
		"clean":             true,
		"dirty":             true,
		"needs maintenance": true,
	}

	if !validStatuses[cleanliness] {
		return errors.New("invalid cleanliness status")
	}

	return nil
}
