package mail

import (
	"bytes"
	"fmt"
	"mime"
	"net/smtp"
	"path/filepath"

	"github.com/jordan-wright/email"
)

// Config holds the configuration for the MailService.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// Service sends an email to a list of addresses.
type Service struct {
	config Config
}

// Attachment represents an email attachment.
type Attachment struct {
	Name string
	Data []byte
}

// NewService creates a new Service with the given configuration.
func NewService(cfg Config) *Service {
	return &Service{
		config: cfg,
	}
}

// Send sends an email with the given subject, body, and optional attachments to the provided addresses.
func (s *Service) Send(to []string, subject, body string, attachments ...Attachment) error {
	if len(to) == 0 {
		return fmt.Errorf("no recipients provided")
	}
	if s.config.Host == "" {
		return fmt.Errorf("SMTP host is not configured")
	}

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	var auth smtp.Auth
	// Only use authentication if username is provided
	if s.config.Username != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	e := email.NewEmail()
	e.From = s.config.From
	e.To = to
	e.Subject = subject
	e.Text = []byte(body)

	// Attachments
	for _, att := range attachments {
		mimeType := mime.TypeByExtension(filepath.Ext(att.Name))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}

		_, err := e.Attach(bytes.NewReader(att.Data), att.Name, mimeType)
		if err != nil {
			return fmt.Errorf("failed to attach file %s: %w", att.Name, err)
		}
	}

	err := e.Send(addr, auth)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
