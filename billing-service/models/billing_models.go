package models

import "time"

// Payment represents a payment transaction
type Payment struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	PaymentStatus string    `json:"payment_status"`
	PaymentDate   time.Time `json:"payment_date"`
	BookingID     int       `json:"booking_id"` // Add BookingID here
}

// Invoice represents an invoice generated for a booking.
type Invoice struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	BookingID     int       `json:"booking_id"`
	Amount        float64   `json:"amount"`
	PaymentStatus string    `json:"payment_status"`
	InvoiceDate   time.Time `json:"invoice_date"`
	CreatedAt     time.Time `json:"created_at"`
}
