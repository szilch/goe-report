package formatter

import (
	"os"
	"testing"
)

func TestNewPDFFormatter(t *testing.T) {
	f := NewPDFFormatter("test_report.pdf")

	if f == nil {
		t.Error("NewPDFFormatter() returned nil")
	}

	if f.filename != "test_report.pdf" {
		t.Errorf("expected filename 'test_report.pdf', got '%s'", f.filename)
	}
}

func TestPDFFormatter_Format_CreatesPDF(t *testing.T) {
	filename := "test_output.pdf"
	defer os.Remove(filename) // Clean up

	f := NewPDFFormatter(filename)
	data := ReportData{
		MonthName:     "01-2026",
		StartDate:     "01.01.2026",
		EndDate:       "31.01.2026",
		SerialNumber:  "ABC123",
		LicensePlate:  "B-GO 123",
		Mileage:       "50000 km",
		KwhPrice:      0.35,
		TotalSessions: 2,
		TotalEnergy:   20.0,
		TotalPrice:    7.0,
		Sessions: []SessionData{
			{StartDate: "01.01.2026 10:00", EndDate: "01.01.2026 12:00", Duration: "7200", Energy: 10.0, Price: 3.50, RFID: "123"},
			{StartDate: "02.01.2026 14:00", EndDate: "02.01.2026 15:00", Duration: "3600", Energy: 10.0, Price: 3.50, RFID: "123"},
		},
	}

	err := f.Format(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check if file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("PDF file was not created")
	}

	// Check file is not empty
	info, _ := os.Stat(filename)
	if info.Size() == 0 {
		t.Error("PDF file is empty")
	}
}

func TestPDFFormatter_Format_NoSessions(t *testing.T) {
	filename := "test_no_sessions.pdf"
	defer os.Remove(filename)

	f := NewPDFFormatter(filename)
	data := ReportData{
		MonthName:     "01-2026",
		StartDate:     "01.01.2026",
		EndDate:       "31.01.2026",
		TotalSessions: 0,
		Sessions:      nil,
	}

	err := f.Format(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check if file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("PDF file was not created")
	}
}

func TestPDFFormatter_Format_NoLicensePlate(t *testing.T) {
	filename := "test_no_license.pdf"
	defer os.Remove(filename)

	f := NewPDFFormatter(filename)
	data := ReportData{
		MonthName:     "01-2026",
		StartDate:     "01.01.2026",
		EndDate:       "31.01.2026",
		LicensePlate:  "",
		TotalSessions: 1,
		Sessions: []SessionData{
			{StartDate: "01.01.2026 10:00", EndDate: "01.01.2026 12:00", Duration: "7200", Energy: 10.0, Price: 3.50, RFID: "123"},
		},
	}

	err := f.Format(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check if file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("PDF file was not created")
	}
}

func TestPDFFormatter_Format_UmlautsInContent(t *testing.T) {
	filename := "test_umlauts.pdf"
	defer os.Remove(filename)

	f := NewPDFFormatter(filename)
	data := ReportData{
		MonthName:     "01-2026",
		StartDate:     "01.01.2026",
		EndDate:       "31.01.2026",
		LicensePlate:  "MÜ-BC 123", // Contains Umlaut
		Mileage:       "50000 km",
		KwhPrice:      0.35,
		TotalSessions: 1,
		Sessions: []SessionData{
			{StartDate: "01.01.2026 10:00", EndDate: "01.01.2026 12:00", Duration: "7200", Energy: 10.0, Price: 3.50, RFID: "123"},
		},
	}

	err := f.Format(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPDFFormatter_Format_ManySessions(t *testing.T) {
	filename := "test_many_sessions.pdf"
	defer os.Remove(filename)

	f := NewPDFFormatter(filename)

	// Create many sessions to test pagination
	var sessions []SessionData
	for i := 0; i < 50; i++ {
		sessions = append(sessions, SessionData{
			StartDate: "01.01.2026 10:00",
			EndDate:   "01.01.2026 12:00",
			Duration:  "7200",
			Energy:    10.0,
			Price:     3.50,
			RFID:      "123",
		})
	}

	data := ReportData{
		MonthName:     "01-2026",
		StartDate:     "01.01.2026",
		EndDate:       "31.01.2026",
		TotalSessions: 50,
		TotalEnergy:   500.0,
		TotalPrice:    175.0,
		Sessions:      sessions,
	}

	err := f.Format(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check if file was created
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("PDF file was not created")
	}
}

func TestPDFFormatter_ImplementsFormatter(t *testing.T) {
	var _ Formatter = &PDFFormatter{}
}

func TestPDFFormatter_InvalidPath(t *testing.T) {
	// Try to write to an invalid path
	f := NewPDFFormatter("/nonexistent/directory/test.pdf")
	data := ReportData{
		MonthName:     "01-2026",
		TotalSessions: 0,
	}

	err := f.Format(data)
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}
