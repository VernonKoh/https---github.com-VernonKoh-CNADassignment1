package handlers

import (
	"cnad_assignment/user-service/database"
	"cnad_assignment/user-service/models"
	"cnad_assignment/user-service/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func ValidateEmail(email string) bool {
	re := regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}$`)
	return re.MatchString(email)
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	log.Println("RegisterUser handler called")

	var user models.User

	// Decode JSON input
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("Error decoding input: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid input"})
		return
	}

	// Debugging received data
	log.Printf("Received user data: Email=%s, Name=%s, Password=%s", user.Email, user.Name, user.Password)

	// Validate email
	if !utils.ValidateEmail(user.Email) {
		log.Println("Invalid email format")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid email address"})
		return
	}

	// Validate password
	if strings.TrimSpace(user.Password) == "" {
		log.Println("Password is required")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Password is required"})
		return
	}

	// Validate name
	if strings.TrimSpace(user.Name) == "" {
		log.Println("Name is required")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Name is required"})
		return
	}

	// Hash the password
	log.Println("Hashing the password")
	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error hashing password"})
		return
	}
	log.Printf("Hashed password: %s", hashedPassword)

	// Assign hashed password to the user
	user.Password = hashedPassword

	// Generate a verification token
	log.Println("Generating verification token")
	token, err := utils.GenerateVerificationToken()
	if err != nil {
		log.Printf("Error generating verification token: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error generating verification token"})
		return
	}

	// Insert the user into the database
	query := "INSERT INTO users (email, password, name, role, verification_token) VALUES (?, ?, ?, ?, ?)"
	_, err = database.DB.Exec(query, user.Email, user.Password, user.Name, user.Role, token)
	if err != nil {
		log.Printf("Error inserting user into database: %v", err)
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(err.Error(), "Duplicate entry") {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{"error": "Email already registered"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to register user"})
		}
		return
	}

	// Simulate sending the email
	verificationLink := fmt.Sprintf("http://localhost:8081/api/v1/users/verify?token=%s", token)
	log.Printf("Send email to %s with verification link: %s", user.Email, verificationLink)

	// Send the actual verification email
	if err := utils.SendVerificationEmail(user.Email, verificationLink); err != nil {
		log.Printf("Error sending verification email: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to send verification email"})
		return
	}

	// Success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered successfully. Please verify your email.",
		"email":   user.Email,
		"token":   token, // Optional: Include token for testing purposes
	})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid input"})
		return
	}

	// Fetch user from the database
	var user models.User
	query := "SELECT id, name, password, is_verified FROM users WHERE email = ?"
	err := database.DB.QueryRow(query, credentials.Email).Scan(&user.ID, &user.Name, &user.Password, &user.IsVerified)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid email or password"})
		return
	}
	log.Printf("Fetched user: ID=%d, Name=%s, Password=%s, IsVerified=%t", user.ID, user.Name, user.Password, user.IsVerified)
	// Check if user is verified
	if !user.IsVerified {
		log.Printf("Email not verified for user: %s", credentials.Email)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Email not verified"})
		return
	}
	log.Printf("Login successful: UserID=%d, Name=%s", user.ID, user.Name)
	// Validate the password
	if !utils.CheckPasswordHash(credentials.Password, user.Password) {
		log.Printf("Invalid password for user: %s", credentials.Email)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid email or password"})
		return
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		log.Printf("Error generating JWT: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to generate token"})
		return
	}

	// Respond with token, userID, and name
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"token":  token,
		"userID": user.ID,
		"name":   user.Name,
	})
}

func VerifyUser(w http.ResponseWriter, r *http.Request) {
	// Extract the token from the query parameters
	token := r.URL.Query().Get("token")
	if token == "" {
		log.Println("Missing token in query string")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Missing token"})
		return
	}

	// Debug: Log the received token
	log.Printf("Verification token received: %s", token)

	// Check if the token exists in the database
	var userID int
	err := database.DB.QueryRow("SELECT id FROM users WHERE verification_token = ?", token).Scan(&userID)
	if err != nil {
		log.Printf("Error finding user with token: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token or user not found"})
		return
	}

	// Debug: Log the user ID associated with the token
	log.Printf("User ID associated with token: %d", userID)

	// Update the user's verification status
	result, err := database.DB.Exec("UPDATE users SET is_verified = TRUE, verification_token = NULL WHERE id = ?", userID)
	if err != nil {
		log.Printf("Error updating user verification status: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Error verifying user"})
		return
	}

	// Check if any row was affected
	rowsAffected, _ := result.RowsAffected()
	log.Printf("Rows affected: %d", rowsAffected)
	if rowsAffected == 0 {
		log.Println("Invalid token or user not found in database")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid token or user not found"})
		return
	}

	log.Println("User email verified successfully")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Email verified successfully"})
}

func UpdateUserMembership(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"] // Get the user ID from the URL
	if id == "" {
		log.Println("Error: Missing or invalid user ID")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	var request struct {
		Role string `json:"role"` // The new membership tier
	}

	log.Printf("Received membership update request: ID=%s", id) // Debug: Log ID

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("Error decoding request body: %v", err) // Debug: Log error
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid input"})
		return
	}

	// Validate membership role
	validRoles := map[string]bool{"Basic": true, "Premium": true, "VIP": true}
	if !validRoles[request.Role] {
		log.Printf("Invalid membership role: %s", request.Role)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid membership tier"})
		return
	}

	// Convert ID to integer
	userID, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("Error converting user ID to integer: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	log.Printf("Updating membership for user ID=%d to Role=%s", userID, request.Role) // Debug: Log role

	// Update the membership tier in the database
	result, err := database.DB.Exec("UPDATE users SET role = ? WHERE id = ?", request.Role, userID)
	if err != nil {
		log.Printf("Database error during update: %v", err) // Debug: Log error
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to update membership tier"})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("No rows updated for user ID: %d", userID)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found or no change in role"})
		return
	}

	log.Printf("Membership tier updated successfully for user ID=%d", userID) // Debug: Confirm success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Membership tier updated successfully", "updatedRole": request.Role})
}

func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"] // Extract user ID from the URL

	// Validate if ID is numeric
	userID, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("Invalid user ID: %s", id)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid user ID"})
		return
	}

	var user models.User

	// Fetch user data from the database
	query := "SELECT id, email, name, role FROM users WHERE id = ?"
	err = database.DB.QueryRow(query, userID).Scan(&user.ID, &user.Email, &user.Name, &user.Role)
	if err != nil {
		log.Printf("Error fetching user profile for ID=%d: %v", userID, err)
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		return
	}

	log.Printf("Fetched user profile for ID=%d", userID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// UpdateUserProfile allows users to update their details and membership
func UpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"] // Extract user ID from URL
	var updates struct {
		Name string `json:"name"`
		Role string `json:"role"`
	}

	// Decode the request body
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Map role to membership_tier_id
	var membershipTierID int
	err := database.DB.QueryRow("SELECT id FROM membership_tiers WHERE name = ?", updates.Role).Scan(&membershipTierID)
	if err != nil {
		log.Printf("Error finding membership tier ID for role %s: %v", updates.Role, err)
		http.Error(w, "Invalid membership role", http.StatusBadRequest)
		return
	}

	// Update user details in the database
	query := "UPDATE users SET name = ?, role = ?, membership_tier_id = ? WHERE id = ?"
	_, err = database.DB.Exec(query, updates.Name, updates.Role, membershipTierID, id)
	if err != nil {
		log.Printf("Error updating user profile: %v", err)
		http.Error(w, "Failed to update user profile", http.StatusInternalServerError)
		return
	}

	log.Printf("Updated user ID=%s: Name=%s, Role=%s, MembershipTierID=%d", id, updates.Name, updates.Role, membershipTierID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
}

func GetMembershipTiers(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query("SELECT id, name, hourly_rate_discount, priority_access, booking_limit FROM membership_tiers")
	if err != nil {
		log.Printf("Error fetching membership tiers: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch membership tiers"})
		return
	}
	defer rows.Close()

	var tiers []map[string]interface{}
	for rows.Next() {
		var id int
		var name string
		var discount float64
		var priorityAccess bool
		var bookingLimit int
		if err := rows.Scan(&id, &name, &discount, &priorityAccess, &bookingLimit); err != nil {
			log.Printf("Error scanning membership tier row: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to process membership tiers"})
			return
		}
		tiers = append(tiers, map[string]interface{}{
			"id":                 id,
			"name":               name,
			"hourlyRateDiscount": discount,
			"priorityAccess":     priorityAccess,
			"bookingLimit":       bookingLimit,
		})
	}

	if len(tiers) == 0 {
		w.WriteHeader(http.StatusNoContent)
		json.NewEncoder(w).Encode(map[string]string{"message": "No membership tiers found"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tiers)
}

func GetUserMembershipBenefits(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"] // Extract user ID from URL
	if id == "" {
		log.Println("Error: User ID is missing or invalid")
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	log.Printf("Fetching membership benefits for user ID: %s", id)

	var tierName string
	var discount float64
	var priorityAccess bool
	var bookingLimit int

	query := `
        SELECT mt.name, mt.hourly_rate_discount, mt.priority_access, mt.booking_limit
        FROM users u
        JOIN membership_tiers mt ON u.membership_tier_id = mt.id
        WHERE u.id = ?
    `
	err := database.DB.QueryRow(query, id).Scan(&tierName, &discount, &priorityAccess, &bookingLimit)
	if err != nil {
		log.Printf("Error fetching membership benefits for user ID=%s: %v", id, err)
		http.Error(w, "User or membership tier not found", http.StatusNotFound)
		return
	}

	log.Printf("Fetched membership benefits for user ID=%s: Tier=%s", id, tierName)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tierName":           tierName,
		"hourlyRateDiscount": discount,
		"priorityAccess":     priorityAccess,
		"bookingLimit":       bookingLimit,
	})
}

func GetUserRentalHistory(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	rows, err := database.DB.Query("SELECT date, vehicle, duration FROM rentals WHERE user_id = ?", id)
	if err != nil {
		log.Printf("Error fetching rental history for user ID=%s: %v", id, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch rental history"})
		return
	}
	defer rows.Close()

	var rentals []map[string]interface{}
	for rows.Next() {
		var date string
		var vehicle string
		var duration int
		if err := rows.Scan(&date, &vehicle, &duration); err != nil {
			log.Printf("Error scanning rental history row: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to process rental history"})
			return
		}
		rentals = append(rentals, map[string]interface{}{
			"date":     date,
			"vehicle":  vehicle,
			"duration": duration,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rentals)
}

func FetchVehiclesFromVehicleService() {
	resp, err := http.Get("http://localhost:8082/api/v1/vehicles")
	if err != nil {
		log.Printf("Error contacting vehicle-service: %v", err)
		return
	}
	defer resp.Body.Close()

	var vehicles []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&vehicles); err != nil {
		log.Printf("Error decoding response: %v", err)
		return
	}

	log.Printf("Fetched vehicles: %+v", vehicles)
}
