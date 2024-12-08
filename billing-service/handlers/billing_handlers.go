package handlers

import (
	"cnad_assignment/billing-service/database" // Import the database package
	"cnad_assignment/billing-service/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Calculate the cost including discount based on user role
func calculateBillingWithDiscount(userID int, startTime, endTime time.Time) (float64, float64, float64, error) {
	// Fetch user's role from the database
	var userRole string
	query := "SELECT role FROM users WHERE id = ?"
	err := database.DB.QueryRow(query, userID).Scan(&userRole) // Access DB from the database package
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to fetch user role: %v", err)
	}

	// Set discount based on membership tier
	var discount float64
	switch userRole {
	case "VIP":
		discount = 0.20 // 20% discount for VIP
	case "Premium":
		discount = 0.10 // 10% discount for Premium
	default:
		discount = 0.00 // No discount for Basic
	}

	// Calculate rental duration in hours
	rentalDuration := endTime.Sub(startTime).Hours()

	// Base rate per hour (example)
	baseRate := 20.00

	// Calculate the cost before discount
	costBeforeDiscount := baseRate * rentalDuration

	// Apply discount
	discountAmount := costBeforeDiscount * discount
	finalCost := costBeforeDiscount - discountAmount

	return costBeforeDiscount, discountAmount, finalCost, nil
}

// FetchBillingDetails fetches the billing details for a user including booking and vehicle details
func FetchBillingDetails(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Fetch the bookings and vehicle details for the user
	bookings, err := database.FetchBookingsByUser(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching bookings: %v", err), http.StatusInternalServerError)
		return
	}

	// Fetch the user's membership role and calculate discount
	var userRole string
	query := "SELECT role FROM users WHERE id = ?"
	err = database.DB.QueryRow(query, userID).Scan(&userRole)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching user role: %v", err), http.StatusInternalServerError)
		return
	}

	// Set discount based on membership tier

	var discountPercentage string
	switch userRole {
	case "VIP":

		discountPercentage = "20%"
	case "Premium":

		discountPercentage = "10%"
	default:

		discountPercentage = "0%"
	}

	// Calculate total cost and gather billing details
	var totalCost float64
	var billingDetails []map[string]interface{}

	// Process each booking
	for _, booking := range bookings {
		// Get the vehicle details
		vehicle := fmt.Sprintf("%s %s (%s)", booking["make"], booking["model"], booking["registration_number"])

		// Calculate the cost for the booking
		costBeforeDiscount, discountAmount, finalCost, err := calculateBillingWithDiscount(userID, booking["start_time"].(time.Time), booking["end_time"].(time.Time))
		if err != nil {
			http.Error(w, fmt.Sprintf("Error calculating billing: %v", err), http.StatusInternalServerError)
			return
		}

		// Add the billing details for each booking to the response
		billingDetails = append(billingDetails, map[string]interface{}{
			"vehicle":              vehicle,
			"start_time":           booking["start_time"],
			"end_time":             booking["end_time"],
			"duration":             booking["end_time"].(time.Time).Sub(booking["start_time"].(time.Time)).Hours(),
			"cost_before_discount": fmt.Sprintf("$%.2f", costBeforeDiscount),
			"discount":             fmt.Sprintf("$%.2f (%s)", discountAmount, discountPercentage),
			"final_cost":           fmt.Sprintf("$%.2f", finalCost),
		})

		// Add the cost to the total cost
		totalCost += finalCost
	}

	// Respond with the total cost, billing details, and user's role and discount percentage
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_cost":      totalCost,
		"billing_details": billingDetails,
		"user_role":       userRole,           // Send the user's role
		"discount":        discountPercentage, // Send the discount percentage
	})
}

// Payment struct
type Payment struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"` // Ensure this is int, not string
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	PaymentStatus string    `json:"payment_status"`
	PaymentDate   time.Time `json:"payment_date"`
	BookingID     int       `json:"booking_id"`
}

// HandlePaymentConfirmation handles payment confirmation, clears the debt, and sends the invoice
func HandlePaymentConfirmation(w http.ResponseWriter, r *http.Request) {
	var paymentDetails struct {
		UserID    int     `json:"user_id"`    // Expecting integer for user_id
		Amount    float64 `json:"amount"`     // Expecting float64 for amount
		BookingID int     `json:"booking_id"` // Expecting integer for booking_id
	}

	// Decode incoming payment details
	if err := json.NewDecoder(r.Body).Decode(&paymentDetails); err != nil {
		log.Printf("Error decoding payment details: %v", err)
		http.Error(w, "Error decoding payment details", http.StatusBadRequest)
		return
	}

	// Log the decoded payment details for debugging
	log.Printf("Received payment details: %+v\n", paymentDetails)

	if paymentDetails.Amount <= 0 {
		http.Error(w, "Invalid payment amount", http.StatusBadRequest)
		return
	}

	// Ensure booking ID is valid and exists
	if paymentDetails.BookingID == 0 {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	// Now you can process the payment
	query := "UPDATE payments SET payment_status = 'paid' WHERE user_id = ? AND payment_status = 'pending' AND booking_id = ?"
	_, err := database.DB.Exec(query, paymentDetails.UserID, paymentDetails.BookingID)
	if err != nil {
		http.Error(w, "Error updating payment status", http.StatusInternalServerError)
		return
	}

	// Generate the invoice
	invoiceID := generateInvoice(paymentDetails.UserID, paymentDetails.BookingID, paymentDetails.Amount)
	if invoiceID == 0 {
		http.Error(w, "Failed to generate invoice", http.StatusInternalServerError)
		return
	}

	// Fetch user email from the database
	var userEmail string
	err = database.DB.QueryRow("SELECT email FROM users WHERE id = ?", paymentDetails.UserID).Scan(&userEmail)
	if err != nil {
		http.Error(w, "Failed to fetch user email", http.StatusInternalServerError)
		return
	}

	// Generate email content
	invoiceDetails := map[string]interface{}{
		"invoice_id": invoiceID,
		"user_id":    paymentDetails.UserID,
		"amount":     paymentDetails.Amount,
		"status":     "Paid",
		"date":       time.Now().Format(time.RFC1123),
	}
	emailContent := utils.GenerateInvoiceEmail(invoiceDetails)

	// Send the invoice email
	err = utils.SendEmail(userEmail, "Your Invoice", emailContent)
	if err != nil {
		http.Error(w, "Failed to send invoice email", http.StatusInternalServerError)
		return
	}

	// Respond to the client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Payment confirmed and invoice sent via email.",
	})
}

// ConfirmPayment confirms the payment and processes the request
func ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	var payment Payment
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding payment details: %v", err), http.StatusBadRequest)
		return
	}

	// Log received data for debugging
	log.Printf("Received payment details: %+v\n", payment)

	// Validate required fields
	if payment.UserID == 0 || payment.Amount <= 0 || payment.PaymentMethod == "" || payment.BookingID == 0 {
		http.Error(w, "Missing required payment information.", http.StatusBadRequest)
		return
	}

	// Process the payment (e.g., Stripe or PayPal) - here we assume it is successful
	payment.PaymentStatus = "completed" // Update this based on payment gateway response
	payment.PaymentDate = time.Now()

	// Store the payment in the database
	query := `
        INSERT INTO payments (user_id, amount, payment_status, payment_method, payment_date, booking_id)
        VALUES (?, ?, ?, ?, ?, ?)
    `
	_, err := database.DB.Exec(query, payment.UserID, payment.Amount, payment.PaymentStatus, payment.PaymentMethod, payment.PaymentDate, payment.BookingID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error processing payment: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate the invoice
	invoiceID := generateInvoice(payment.UserID, payment.BookingID, payment.Amount)
	if invoiceID == 0 {
		http.Error(w, "Failed to generate invoice", http.StatusInternalServerError)
		return
	}

	// Send the invoice via email
	invoiceDetails := map[string]interface{}{
		"invoice_id": invoiceID,
		"user_id":    payment.UserID,
		"amount":     payment.Amount,
		"status":     "Paid",
		"date":       time.Now().Format(time.RFC1123),
	}

	emailContent := utils.GenerateInvoiceEmail(invoiceDetails)

	// Send the invoice email
	userEmail := "" // Fetch the user email based on payment.UserID
	err = utils.SendEmail(userEmail, "Your Invoice", emailContent)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send invoice email: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond to the client
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Payment processed successfully and invoice sent.",
	})
}

// generateInvoice creates a new invoice in the existing table and returns the invoice ID.
func generateInvoice(userID int, bookingID int, amount float64) int {
	// Ensure the booking_id exists in the bookings table
	var validBookingID int
	err := database.DB.QueryRow("SELECT id FROM bookings WHERE id = ?", bookingID).Scan(&validBookingID)
	if err != nil {
		fmt.Printf("Invalid booking ID: %v\n", err)
		return 0
	}

	// Insert invoice if booking ID is valid
	query := `
        INSERT INTO invoices (user_id, booking_id, amount, payment_status, invoice_date)
        VALUES (?, ?, ?, 'Paid', ?)
    `
	result, err := database.DB.Exec(query, userID, bookingID, amount, time.Now())
	if err != nil {
		fmt.Printf("Error generating invoice: %v\n", err)
		return 0
	}

	invoiceID, _ := result.LastInsertId()
	return int(invoiceID)
}
