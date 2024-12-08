package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

var jwtSecret = []byte("your_jwt_secret_key") // Ensure this line has no issues.

// GenerateJWT generates a JWT token for authenticated users
func GenerateJWT(userID int) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,                                // Subject (user ID)
		"exp": time.Now().Add(24 * time.Hour).Unix(), // Expiry time (24 hours)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// HashPassword encrypts a plain password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash verifies a hashed password
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateVerificationToken creates a random token for email verification
func GenerateVerificationToken() (string, error) {
	token := make([]byte, 16)
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}

// ValidateEmail checks if the email address is valid
func ValidateEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.(com|org|net|edu)$`)
	return re.MatchString(email)
}

// SMTPConfig holds SMTP server configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

// DefaultSMTPConfig for your email server
var DefaultSMTPConfig = SMTPConfig{
	Host:     "smtp.gmail.com", // Change this based on your SMTP provider
	Port:     587,
	Username: "checkme123ymail.com@gmail.com", // Your email address
	Password: "kket mgmy ymea ywuz",           // Your email password or app password
}

// SendVerificationEmail sends an email with the verification link
func SendVerificationEmail(to, verificationLink string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", DefaultSMTPConfig.Username)
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Email Verification")
	m.SetBody("text/plain", fmt.Sprintf("Please verify your email by clicking the link: %s", verificationLink))

	dialer := gomail.NewDialer(DefaultSMTPConfig.Host, DefaultSMTPConfig.Port, DefaultSMTPConfig.Username, DefaultSMTPConfig.Password)
	return dialer.DialAndSend(m)
}

// ValidateJWT validates the JWT token from the Authorization header and returns the user ID.
func ValidateJWT(r *http.Request) (int, error) {
	// Get the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0, errors.New("missing authorization header")
	}

	// Check if the header starts with "Bearer "
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0, errors.New("invalid authorization header format")
	}

	// Parse the token
	tokenString := parts[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return 0, err
	}

	// Extract claims and validate
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userID, ok := claims["sub"].(float64); ok {
			return int(userID), nil
		}
	}

	return 0, errors.New("invalid token")
}
