package email

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
)

type EmailService struct {
	From     string
	Password string
	SMTPHost string
	SMTPPort string
}

// NewEmailService creates a new email service instance
func NewEmailService() *EmailService {
	return &EmailService{
		From:     os.Getenv("SMTP_EMAIL"),
		Password: os.Getenv("SMTP_PASSWORD"),
		SMTPHost: os.Getenv("SMTP_HOST"),
		SMTPPort: os.Getenv("SMTP_PORT"),
	}
}

// SendVerificationEmail sends an email verification link to the user
func (es *EmailService) SendVerificationEmail(to, username, token string) error {
	if es.From == "" || es.Password == "" {
		log.Println("⚠️  SMTP credentials not configured, skipping email send")
		return fmt.Errorf("SMTP credentials not configured")
	}

	// Get base URL from environment
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", baseURL, token)

	subject := "Verify Your Email - AIManage"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4F46E5; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background-color: #f9f9f9; padding: 30px; border-radius: 0 0 5px 5px; }
        .button { display: inline-block; padding: 12px 30px; background-color: #4F46E5; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Welcome to AIManage!</h1>
        </div>
        <div class="content">
            <p>Hi %s,</p>
            <p>Thank you for registering with AIManage! Please verify your email address to complete your registration.</p>
            <p style="text-align: center;">
                <a href="%s" class="button">Verify Email Address</a>
            </p>
            <p>Or copy and paste this link into your browser:</p>
            <p style="word-break: break-all; background-color: #e9ecef; padding: 10px; border-radius: 3px;">%s</p>
            <p>This verification link will expire in 24 hours.</p>
            <p>If you didn't create an account with AIManage, please ignore this email.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 AIManage. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, username, verificationLink, verificationLink)

	// Compose message
	message := []byte(
		"From: " + es.From + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n" +
			"\r\n" +
			body + "\r\n")

	// Set up authentication
	auth := smtp.PlainAuth("", es.From, es.Password, es.SMTPHost)

	// Send email
	addr := es.SMTPHost + ":" + es.SMTPPort
	err := smtp.SendMail(addr, auth, es.From, []string{to}, message)
	if err != nil {
		log.Printf("❌ Failed to send verification email to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("✅ Verification email sent to %s", to)
	return nil
}

// SendWelcomeEmail sends a welcome email after email verification
func (es *EmailService) SendWelcomeEmail(to, username string) error {
	if es.From == "" || es.Password == "" {
		log.Println("⚠️  SMTP credentials not configured, skipping email send")
		return fmt.Errorf("SMTP credentials not configured")
	}

	subject := "Welcome to AIManage!"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4F46E5; color: white; padding: 20px; text-align: center; border-radius: 5px 5px 0 0; }
        .content { background-color: #f9f9f9; padding: 30px; border-radius: 0 0 5px 5px; }
        .footer { text-align: center; margin-top: 20px; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Email Verified Successfully!</h1>
        </div>
        <div class="content">
            <p>Hi %s,</p>
            <p>Your email has been verified successfully. You can now access all features of AIManage!</p>
            <p>Get started by:</p>
            <ul>
                <li>Creating your first AI model</li>
                <li>Exploring the community marketplace</li>
                <li>Checking out our documentation</li>
            </ul>
            <p>If you have any questions, feel free to reach out to our support team.</p>
        </div>
        <div class="footer">
            <p>&copy; 2024 AIManage. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, username)

	// Compose message
	message := []byte(
		"From: " + es.From + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n" +
			"\r\n" +
			body + "\r\n")

	// Set up authentication
	auth := smtp.PlainAuth("", es.From, es.Password, es.SMTPHost)

	// Send email
	addr := es.SMTPHost + ":" + es.SMTPPort
	err := smtp.SendMail(addr, auth, es.From, []string{to}, message)
	if err != nil {
		log.Printf("❌ Failed to send welcome email to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("✅ Welcome email sent to %s", to)
	return nil
}
