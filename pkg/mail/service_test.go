package mail

import (
	"testing"
)

func TestNewService(t *testing.T) {
	cfg := Config{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user@example.com",
		Password: "secret",
		From:     "sender@example.com",
	}

	s := NewService(cfg)

	if s == nil {
		t.Error("NewService() returned nil")
	}

	if s.config.Host != "smtp.example.com" {
		t.Errorf("expected Host 'smtp.example.com', got '%s'", s.config.Host)
	}

	if s.config.Port != 587 {
		t.Errorf("expected Port 587, got %d", s.config.Port)
	}

	if s.config.Username != "user@example.com" {
		t.Errorf("expected Username 'user@example.com', got '%s'", s.config.Username)
	}

	if s.config.Password != "secret" {
		t.Errorf("expected Password 'secret', got '%s'", s.config.Password)
	}

	if s.config.From != "sender@example.com" {
		t.Errorf("expected From 'sender@example.com', got '%s'", s.config.From)
	}
}

func TestConfig_Struct(t *testing.T) {
	cfg := Config{
		Host:     "smtp.gmail.com",
		Port:     465,
		Username: "test@gmail.com",
		Password: "password123",
		From:     "noreply@example.com",
	}

	if cfg.Host != "smtp.gmail.com" {
		t.Errorf("expected Host 'smtp.gmail.com', got '%s'", cfg.Host)
	}

	if cfg.Port != 465 {
		t.Errorf("expected Port 465, got %d", cfg.Port)
	}
}

func TestAttachment_Struct(t *testing.T) {
	att := Attachment{
		Name: "report.pdf",
		Data: []byte("PDF content here"),
	}

	if att.Name != "report.pdf" {
		t.Errorf("expected Name 'report.pdf', got '%s'", att.Name)
	}

	if string(att.Data) != "PDF content here" {
		t.Errorf("expected Data 'PDF content here', got '%s'", string(att.Data))
	}
}

func TestService_Send_NoRecipients(t *testing.T) {
	cfg := Config{
		Host: "smtp.example.com",
		Port: 587,
	}
	s := NewService(cfg)

	err := s.Send([]string{}, "Subject", "Body")

	if err == nil {
		t.Error("expected error for empty recipients, got nil")
	}

	if err.Error() != "no recipients provided" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestService_Send_NoHost(t *testing.T) {
	cfg := Config{
		Host: "",
		Port: 587,
	}
	s := NewService(cfg)

	err := s.Send([]string{"recipient@example.com"}, "Subject", "Body")

	if err == nil {
		t.Error("expected error for missing host, got nil")
	}

	if err.Error() != "SMTP host is not configured" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestService_Send_WithAttachment(t *testing.T) {
	cfg := Config{
		Host: "", // Empty host to trigger error early after attachment processing
		Port: 587,
	}
	s := NewService(cfg)

	att := Attachment{
		Name: "test.pdf",
		Data: []byte("test content"),
	}

	err := s.Send([]string{"recipient@example.com"}, "Subject", "Body", att)

	// Should fail at SMTP host check, not at attachment
	if err == nil {
		t.Error("expected error, got nil")
	}

	if err.Error() != "SMTP host is not configured" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestService_Send_MultipleAttachments(t *testing.T) {
	cfg := Config{
		Host: "",
		Port: 587,
	}
	s := NewService(cfg)

	att1 := Attachment{Name: "file1.pdf", Data: []byte("content1")}
	att2 := Attachment{Name: "file2.pdf", Data: []byte("content2")}

	err := s.Send([]string{"recipient@example.com"}, "Subject", "Body", att1, att2)

	// Should fail at SMTP host check
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestService_Send_NilRecipients(t *testing.T) {
	cfg := Config{
		Host: "smtp.example.com",
		Port: 587,
	}
	s := NewService(cfg)

	err := s.Send(nil, "Subject", "Body")

	if err == nil {
		t.Error("expected error for nil recipients, got nil")
	}

	if err.Error() != "no recipients provided" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestConfig_DefaultValues(t *testing.T) {
	cfg := Config{}

	if cfg.Host != "" {
		t.Errorf("expected empty Host, got '%s'", cfg.Host)
	}

	if cfg.Port != 0 {
		t.Errorf("expected Port 0, got %d", cfg.Port)
	}

	if cfg.Username != "" {
		t.Errorf("expected empty Username, got '%s'", cfg.Username)
	}
}
