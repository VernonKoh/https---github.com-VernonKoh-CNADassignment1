package models

import "time"

type Vehicle struct {
	ID                 int       `json:"id"`
	Make               string    `json:"make"`
	Model              string    `json:"model"`
	RegistrationNumber string    `json:"registration_number"`
	IsAvailable        bool      `json:"is_available"`
	CreatedAt          time.Time `json:"created_at"`
}

type Booking struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	VehicleID int       `json:"vehicle_id"`
	StartTime time.Time `json:"start_time"` // Updated to time.Time
	EndTime   time.Time `json:"end_time"`   // Updated to time.Time
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type VehicleStatus struct {
	VehicleID   int    `json:"vehicle_id"`
	Location    string `json:"location"`
	ChargeLevel int    `json:"charge_level"`
	Cleanliness string `json:"cleanliness"`
	UpdatedAt   string `json:"updated_at"`
}
