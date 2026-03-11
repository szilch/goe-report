package mail

import (
	"bytes"
	"fmt"
	"goe-report/pkg/config"
	"mime"
	"net/smtp"
	"path/filepath"

	"github.com/jordan-wright/email"
	"github.com/spf13/viper"
)

// Service sends an email to a list of addresses.
type Service struct {
	host     string
	port     int
	username string
	password string
	from     string
}

// Attachment represents an email attachment.
type Attachment struct {
	Name string
	Data []byte
}

// NewService creates a new Service, fetching its configuration directly.
func NewService() *Service {
	return &Service{
		host:     viper.GetString(config.KeyMailHost),
		port:     viper.GetInt(config.KeyMailPort),
		username: viper.GetString(config.KeyMailUsername),
		password: viper.GetString(config.KeyMailPassword),
		from:     viper.GetString(config.KeyMailFrom),
	}
}

// Send sends an email with the given subject, body, and optional attachments to the provided addresses.
func (s *Service) Send(to []string, subject, body string, attachments ...Attachment) error {
	if len(to) == 0 {
		return fmt.Errorf("no recipients provided")
	}
	if s.host == "" {
		return fmt.Errorf("SMTP host is not configured")
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	var auth smtp.Auth
	// Only use authentication if username is provided
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	e := email.NewEmail()
	e.From = s.from
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
