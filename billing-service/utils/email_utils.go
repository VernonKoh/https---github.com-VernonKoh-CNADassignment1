package utils

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

// SMTPConfig holds the SMTP server configuration.
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

// DefaultSMTPConfig for sending emails.
var DefaultSMTPConfig = SMTPConfig{
	Host:     "smtp.gmail.com", // Change to your SMTP provider
	Port:     587,
	Username: "checkme123ymail.com@gmail.com", // Replace with your email
	Password: "kket mgmy ymea ywuz",           // Replace with your app password
}

// SendEmail sends an email to the specified recipient with the given subject and body.
func SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", DefaultSMTPConfig.Username)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body) // HTML body

	dialer := gomail.NewDialer(DefaultSMTPConfig.Host, DefaultSMTPConfig.Port, DefaultSMTPConfig.Username, DefaultSMTPConfig.Password)
	return dialer.DialAndSend(m)
}

// GenerateInvoiceEmail generates the HTML content for the invoice.
func GenerateInvoiceEmail(invoiceDetails map[string]interface{}) string {
	return fmt.Sprintf(`
		<h1>Invoice</h1>
		<p><strong>Invoice ID:</strong> %d</p>
		<p><strong>User ID:</strong> %d</p>
		<p><strong>Amount:</strong> $%.2f</p>
		<p><strong>Status:</strong> %s</p>
		<p><strong>Date:</strong> %s</p>
	`,
		invoiceDetails["invoice_id"],
		invoiceDetails["user_id"],
		invoiceDetails["amount"],
		invoiceDetails["status"],
		invoiceDetails["date"],
	)
}
