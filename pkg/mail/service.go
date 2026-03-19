package mail

import (
	"bytes"
	"fmt"
	"mime"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"

	"echarge-report/pkg/models"

	"github.com/jordan-wright/email"
)

// Config specifies the connection and sender details for the mail service.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       string
}

// Service is responsible for sending emails with attachments via SMTP.
type Service struct {
	cfg    Config
	sendFn func(to []string, subject, body string, attachments ...Attachment) error
}

// Attachment represents a file to be attached to an email.
type Attachment struct {
	Name string
	Data []byte
}

// NewService creates a new mail Service using the provided configuration.
func NewService(cfg Config) *Service {
	return &Service{
		cfg: cfg,
	}
}

func (s *Service) buildEmail(to []string, subject, body string, attachments ...Attachment) (*email.Email, error) {
	e := email.NewEmail()
	e.From = s.cfg.From
	e.To = to
	e.Subject = subject
	e.Text = []byte(body)

	for _, att := range attachments {
		mimeType := mime.TypeByExtension(filepath.Ext(att.Name))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		_, err := e.Attach(bytes.NewReader(att.Data), att.Name, mimeType)
		if err != nil {
			return nil, fmt.Errorf("attach file %s: %w", att.Name, err)
		}
	}
	return e, nil
}

// Send builds and sends an email to the specified recipients.
func (s *Service) Send(to []string, subject, body string, attachments ...Attachment) error {
	if s.sendFn != nil {
		return s.sendFn(to, subject, body, attachments...)
	}

	if len(to) == 0 {
		return fmt.Errorf("no recipients provided")
	}
	if s.cfg.Host == "" {
		return fmt.Errorf("SMTP host is not configured")
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}

	e, err := s.buildEmail(to, subject, body, attachments...)
	if err != nil {
		return err
	}

	if err := e.Send(addr, auth); err != nil {
		return fmt.Errorf("send email via SMTP: %w", err)
	}
	return nil
}

// SendReportEmail reads the generated PDF report and sends it to the configured recipients.
func (s *Service) SendReportEmail(reportFile string, data models.ReportData) error {
	toRaw := s.cfg.To
	if toRaw == "" {
		return fmt.Errorf("cannot send email because 'mail_to' is not configured")
	}

	var recipients []string
	for _, r := range strings.Split(toRaw, ",") {
		if trimmed := strings.TrimSpace(r); trimmed != "" {
			recipients = append(recipients, trimmed)
		}
	}

	if len(recipients) == 0 {
		return fmt.Errorf("no valid recipient addresses found in 'mail_to' configuration")
	}

	pdfData, err := os.ReadFile(reportFile)
	if err != nil {
		return fmt.Errorf("read generated PDF for email attachment: %w", err)
	}

	subject := fmt.Sprintf("Ladebericht - %s (%s)", data.LicensePlate, data.PeriodLabel)
	body := fmt.Sprintf("Hallo,\n\nangehängt findest du den Ladebericht für das Kennzeichen %s für den Zeitraum %s.\n\nViele Grüße,\necharge-report", data.LicensePlate, data.PeriodLabel)

	attachment := Attachment{
		Name: reportFile,
		Data: pdfData,
	}

	if err := s.Send(recipients, subject, body, attachment); err != nil {
		return fmt.Errorf("send email to recipients: %w", err)
	}
	return nil
}
