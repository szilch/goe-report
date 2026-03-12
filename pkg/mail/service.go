package mail

import (
	"bytes"
	"echarge-report/pkg/config"
	"echarge-report/pkg/models"
	"fmt"
	"mime"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"

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
	// sendFn allows overriding the actual send call in tests.
	sendFn func(to []string, subject, body string, attachments ...Attachment) error
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

// buildEmail constructs an email.Email from parameters, attaching any provided files.
func (s *Service) buildEmail(to []string, subject, body string, attachments ...Attachment) (*email.Email, error) {
	e := email.NewEmail()
	e.From = s.from
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
			return nil, fmt.Errorf("failed to attach file %s: %w", att.Name, err)
		}
	}
	return e, nil
}

// Send sends an email with the given subject, body, and optional attachments to the provided addresses.
func (s *Service) Send(to []string, subject, body string, attachments ...Attachment) error {
	if s.sendFn != nil {
		return s.sendFn(to, subject, body, attachments...)
	}

	if len(to) == 0 {
		return fmt.Errorf("no recipients provided")
	}
	if s.host == "" {
		return fmt.Errorf("SMTP host is not configured")
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	var auth smtp.Auth
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	e, err := s.buildEmail(to, subject, body, attachments...)
	if err != nil {
		return err
	}

	if err := e.Send(addr, auth); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}

// SendReportEmail reads the generated PDF report and sends it via email based on ReportData.
func (s *Service) SendReportEmail(reportFile string, data models.ReportData) error {
	toRaw := viper.GetString(config.KeyMailTo)
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
		return fmt.Errorf("error reading generated PDF for email attachment: %w", err)
	}

	subject := fmt.Sprintf("Ladebericht - %s (%s)", data.LicensePlate, data.PeriodLabel)
	body := fmt.Sprintf("Hallo,\n\nangehängt findest du den Ladebericht für das Kennzeichen %s für den Zeitraum %s.\n\nViele Grüße,\necharge-report", data.LicensePlate, data.PeriodLabel)

	attachment := Attachment{
		Name: reportFile,
		Data: pdfData,
	}

	if err := s.Send(recipients, subject, body, attachment); err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}
	return nil
}
