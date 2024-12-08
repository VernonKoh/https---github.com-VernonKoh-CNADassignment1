package models

import "time"

// User represents a user in the system
type User struct {
	ID                int       `json:"id"`
	Email             string    `json:"email"`
	Password          string    `json:"password"`
	Name              string    `json:"name"`
	Role              string    `json:"role"`
	CreatedAt         time.Time `json:"created_at"`
	IsVerified        bool      `json:"is_verified"`        // Add this field
	VerificationToken string    `json:"verification_token"` // Add this field if needed for verification handling
}
