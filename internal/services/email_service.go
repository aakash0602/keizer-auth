package services

import (
	"fmt"
	"net/smtp"
	"os"
)

var (
	smtpHost     = os.Getenv("SMTP_HOST")
	smtpPort     = os.Getenv("SMTP_PORT")
	smtpUser     = os.Getenv("SMTP_USER")
	smtpPassword = os.Getenv("SMTP_PASSWORD")
	from         = os.Getenv("EMAIL_FROM")
)

type EmailService struct {
	host string
	port string
	user string
	pass string
	from string
}

func NewEmailService() *EmailService {
	return &EmailService{host: smtpHost, port: smtpPort, user: smtpUser, pass: smtpPassword, from: from}
}

func (es *EmailService) SendEmail(to string, subject string, body string) error {
	message := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	var auth smtp.Auth
	if es.pass != "" {
		auth = smtp.PlainAuth("", es.user, es.pass, es.host)
	}
	err := smtp.SendMail(es.host+":"+es.port, auth, es.from, []string{to}, message)

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

func SendOTPEmail(to string, otp string) error {
	emailService := NewEmailService()
	err := emailService.SendEmail(to, "OTP Verification", "Your OTP is "+otp)
	if err != nil {
		return err
	}
	return nil
}