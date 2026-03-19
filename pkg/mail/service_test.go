package mail

import (
	"bytes"
	"echarge-report/pkg/config"
	"echarge-report/pkg/models"
	"mime"
	"mime/multipart"
	"net/mail"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

var smtpSendFunc = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	return smtp.SendMail(addr, a, from, to, msg)
}

func newTestService() *Service {
	return &Service{
		cfg: Config{
			Host:     "localhost",
			Port:     1025,
			Username: "",
			Password: "",
			From:     "test@example.com",
		},
	}
}

func TestService_Send_NoRecipients(t *testing.T) {
	s := newTestService()
	err := s.Send([]string{}, "Test Subject", "Test Body")
	if err == nil {
		t.Fatalf("Expected error for no recipients, got nil")
	}
}

func TestService_Send_NoHost(t *testing.T) {
	s := &Service{cfg: Config{Host: "", Port: 1025, From: "a@b.com"}}
	err := s.Send([]string{"to@example.com"}, "Subject", "Body")
	if err == nil {
		t.Fatalf("Expected error for missing host, got nil")
	}
}

func TestAttachment_MimeType(t *testing.T) {
	tests := []struct {
		filename     string
		expectedMime string
	}{
		{"report.pdf", "application/pdf"},
		{"archive.zip", "application/zip"},
		{"file.unknownext", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			detected := mime.TypeByExtension(filepath.Ext(tt.filename))
			if detected == "" {
				detected = "application/octet-stream"
			}
			mediaType, _, _ := mime.ParseMediaType(detected)
			expectedMediaType, _, _ := mime.ParseMediaType(tt.expectedMime)
			if mediaType != expectedMediaType {
				t.Errorf("Expected MIME type %s for %s, got %s", expectedMediaType, tt.filename, mediaType)
			}
		})
	}
}

func TestService_buildEmail_Attachment(t *testing.T) {
	s := newTestService()

	att := Attachment{
		Name: "report.pdf",
		Data: []byte("%PDF-1.4 fake pdf content"),
	}

	e, err := s.buildEmail([]string{"to@example.com"}, "Test Subject", "Test Body", att)
	if err != nil {
		t.Fatalf("buildEmail failed: %v", err)
	}

	if e.Subject != "Test Subject" {
		t.Errorf("Expected subject 'Test Subject', got '%s'", e.Subject)
	}
	if len(e.Attachments) != 1 {
		t.Errorf("Expected 1 attachment, got %d", len(e.Attachments))
	}
	if e.Attachments[0].Filename != "report.pdf" {
		t.Errorf("Expected attachment named 'report.pdf', got '%s'", e.Attachments[0].Filename)
	}
}

func TestService_SendReportEmail_NoMailTo(t *testing.T) {
	viper.Set(config.KeyMailTo, "")
	defer viper.Reset()

	s := newTestService()
	err := s.SendReportEmail("report.pdf", models.ReportData{})
	if err == nil {
		t.Fatalf("Expected error for missing mail_to, got nil")
	}
}

func TestService_SendReportEmail_InvalidFile(t *testing.T) {
	viper.Set(config.KeyMailTo, "to@example.com")
	defer viper.Reset()

	s := newTestService()
	err := s.SendReportEmail("/nonexistent/path/report.pdf", models.ReportData{})
	if err == nil {
		t.Fatalf("Expected error for nonexistent file, got nil")
	}
}

func TestService_SendReportEmail_SubjectAndBody(t *testing.T) {
	viper.Set(config.KeyMailTo, "to@example.com")
	defer viper.Reset()

	tmpFile, err := os.CreateTemp("", "report-*.pdf")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("fake pdf content"))
	tmpFile.Close()

	var capturedSubject, capturedBody string
	var capturedTo []string

	s := &Service{
		cfg: Config{
			Host: "localhost",
			Port: 1025,
			From: "from@example.com",
			To:   "to@example.com",
		},
		sendFn: func(to []string, subject, body string, attachments ...Attachment) error {
			capturedTo = to
			capturedSubject = subject
			capturedBody = body
			return nil
		},
	}

	data := models.ReportData{
		LicensePlate: "W-TEST123",
		PeriodLabel:  "01-2026",
	}

	err = s.SendReportEmail(tmpFile.Name(), data)
	if err != nil {
		t.Fatalf("SendReportEmail failed: %v", err)
	}

	if !strings.Contains(capturedSubject, "W-TEST123") {
		t.Errorf("Expected subject to contain 'W-TEST123', got: %s", capturedSubject)
	}
	if !strings.Contains(capturedSubject, "01-2026") {
		t.Errorf("Expected subject to contain '01-2026', got: %s", capturedSubject)
	}
	if !strings.Contains(capturedBody, "W-TEST123") {
		t.Errorf("Expected body to contain 'W-TEST123', got: %s", capturedBody)
	}
	if len(capturedTo) != 1 || capturedTo[0] != "to@example.com" {
		t.Errorf("Expected recipient to@example.com, got %v", capturedTo)
	}
}

func parseMultipartEmail(t *testing.T, rawEmail []byte) (*mail.Message, *multipart.Reader) {
	t.Helper()
	msg, err := mail.ReadMessage(bytes.NewReader(rawEmail))
	if err != nil {
		t.Fatalf("Failed to parse email: %v", err)
	}
	ct := msg.Header.Get("Content-Type")
	_, params, err := mime.ParseMediaType(ct)
	if err != nil {
		t.Fatalf("Failed to parse content type: %v", err)
	}
	return msg, multipart.NewReader(msg.Body, params["boundary"])
}

func TestService_SendReportEmail_Recipients(t *testing.T) {
	viper.Set(config.KeyMailTo, "a@example.com, b@example.com , c@example.com")
	defer viper.Reset()

	tmpFile, err := os.CreateTemp("", "report-*.pdf")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("fake"))
	tmpFile.Close()

	var capturedTo []string
	s := &Service{
		cfg: Config{
			Host: "localhost",
			Port: 1025,
			From: "from@example.com",
			To:   "a@example.com, b@example.com , c@example.com",
		},
		sendFn: func(to []string, subject, body string, attachments ...Attachment) error {
			capturedTo = to
			return nil
		},
	}

	err = s.SendReportEmail(tmpFile.Name(), models.ReportData{})
	if err != nil {
		t.Fatalf("SendReportEmail failed: %v", err)
	}

	if len(capturedTo) != 3 {
		t.Errorf("Expected 3 recipients, got %d: %v", len(capturedTo), capturedTo)
	}
}

func TestNewService(t *testing.T) {
	cfg := Config{Host: "smtp.test.com", Port: 587}
	s := NewService(cfg)
	if s.cfg.Host != "smtp.test.com" || s.cfg.Port != 587 {
		t.Errorf("NewService failed to store config correctly: %v", s.cfg)
	}
}

func TestService_SendReportEmail_NoLicensePlate(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "report-*.pdf")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("fake"))
	tmpFile.Close()

	var capturedSubject string
	s := &Service{
		cfg: Config{Host: "localhost", Port: 1025, To: "to@example.com"},
		sendFn: func(to []string, subject, body string, attachments ...Attachment) error {
			capturedSubject = subject
			return nil
		},
	}

	data := models.ReportData{PeriodLabel: "02-2026"}
	s.SendReportEmail(tmpFile.Name(), data)

	if !strings.Contains(capturedSubject, "Ladebericht (02-2026)") {
		t.Errorf("Expected subject 'Ladebericht (02-2026)', got: %s", capturedSubject)
	}
}

func TestService_SendReportEmail_EmptyRecipientsAfterSplit(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "report-*.pdf")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte("fake"))
	tmpFile.Close()

	s := &Service{
		cfg: Config{Host: "localhost", Port: 1025, To: ", ,"},
	}

	err := s.SendReportEmail(tmpFile.Name(), models.ReportData{})
	if err == nil {
		t.Error("Expected error for empty/whitespace recipients, got nil")
	}
}

func TestService_Send_AuthUsed(t *testing.T) {
	// Since we can't easily mock e.Send without dependency injection,
	// we at least test the branch where auth is created.
	s := &Service{
		cfg: Config{Host: "localhost", Port: 1025, Username: "user", Password: "pwd"},
		sendFn: func(to []string, subject, body string, attachments ...Attachment) error {
			return nil
		},
	}
	err := s.Send([]string{"to@example.com"}, "Subject", "Body")
	if err != nil {
		t.Errorf("Send failed: %v", err)
	}
}
