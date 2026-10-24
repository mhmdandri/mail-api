package mailer

import (
	"crypto/tls"
	"fmt"

	"mail-api/internal/config"

	"gopkg.in/gomail.v2"
)

type Mailer struct {
	dialer    *gomail.Dialer
	fromEmail string
	fromName  string
}

func New(cfg *config.Config) *Mailer {
	dialer := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)

	// Ensure STARTTLS uses modern TLS and matches the iCloud SMTP host name
	dialer.TLSConfig = &tls.Config{
		ServerName: cfg.SMTPHost,
		MinVersion: tls.VersionTLS12,
	}

	return &Mailer{
		dialer:    dialer,
		fromEmail: cfg.FromEmail,
		fromName:  cfg.FromName,
	}
}

// Send sends an email with optional text and/or HTML parts.
func (m *Mailer) Send(to, subject, text, html string) error {
	msg := gomail.NewMessage()

	// Format Sender / From header: "From Name" <from@domain.com>
	if m.fromName != "" {
		msg.SetAddressHeader("From", m.fromEmail, m.fromName)
	} else {
		msg.SetHeader("From", m.fromEmail)
	}

	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)

	// Multipart MIME handling
	switch {
	case text != "" && html != "":
		// Multipart alternative: Plaintext body + HTML alternative
		msg.SetBody("text/plain", text)
		msg.AddAlternative("text/html", html)
	case html != "":
		msg.SetBody("text/html", html)
	case text != "":
		msg.SetBody("text/plain", text)
	default:
		return fmt.Errorf("pesan email harus memiliki setidaknya konten 'text' atau 'html'")
	}

	if err := m.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("kegagalan pengiriman SMTP: %w", err)
	}

	return nil
}
