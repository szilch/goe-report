package formatter

import (
	"testing"
)

func TestFormatKWhPrice(t *testing.T) {
	tests := []struct {
		name     string
		price    float64
		expected string
	}{
		{"Standard", 0.35, "0,3500 €"},
		{"Four Decimals", 0.3456, "0,3456 €"},
		{"Zero", 0.0, "0,0000 €"},
		{"Large Value", 1.2345, "1,2345 €"},
		{"Small Value", 0.0001, "0,0001 €"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatKWhPrice(tt.price)
			if result != tt.expected {
				t.Errorf("FormatKWhPrice(%f) = %s, want %s", tt.price, result, tt.expected)
			}
		})
	}
}

func TestFormatPrice(t *testing.T) {
	tests := []struct {
		name     string
		price    float64
		expected string
	}{
		{"Standard", 10.50, "10,50 €"},
		{"Two Decimals", 10.99, "10,99 €"},
		{"Zero", 0.0, "0,00 €"},
		{"Large Value", 1234.56, "1.234,56 €"},
		{"Rounded", 10.556, "10,56 €"},
		{"Small Value", 0.01, "0,01 €"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPrice(tt.price)
			if result != tt.expected {
				t.Errorf("FormatPrice(%f) = %s, want %s", tt.price, result, tt.expected)
			}
		})
	}
}

func TestSessionData_Struct(t *testing.T) {
	session := SessionData{
		StartDate: "2026-01-15 10:00:00",
		EndDate:   "2026-01-15 12:00:00",
		Duration:  "7200",
		Energy:    10.5,
		Price:     3.675,
		RFID:      "12345",
	}

	if session.StartDate != "2026-01-15 10:00:00" {
		t.Errorf("expected StartDate '2026-01-15 10:00:00', got '%s'", session.StartDate)
	}

	if session.EndDate != "2026-01-15 12:00:00" {
		t.Errorf("expected EndDate '2026-01-15 12:00:00', got '%s'", session.EndDate)
	}

	if session.Duration != "7200" {
		t.Errorf("expected Duration '7200', got '%s'", session.Duration)
	}

	if session.Energy != 10.5 {
		t.Errorf("expected Energy 10.5, got %.2f", session.Energy)
	}

	if session.Price != 3.675 {
		t.Errorf("expected Price 3.675, got %.3f", session.Price)
	}

	if session.RFID != "12345" {
		t.Errorf("expected RFID '12345', got '%s'", session.RFID)
	}
}

func TestReportData_Struct(t *testing.T) {
	report := ReportData{
		MonthName:     "01-2026",
		StartDate:     "01.01.2026",
		EndDate:       "31.01.2026",
		SerialNumber:  "ABC123",
		LicensePlate:  "B-GO 123",
		Mileage:       "50000 km",
		KwhPrice:      0.35,
		TotalSessions: 5,
		TotalEnergy:   50.0,
		TotalPrice:    17.50,
		Sessions: []SessionData{
			{StartDate: "01.01.2026", EndDate: "01.01.2026", Duration: "3600", Energy: 10.0, Price: 3.50, RFID: "123"},
		},
	}

	if report.MonthName != "01-2026" {
		t.Errorf("expected MonthName '01-2026', got '%s'", report.MonthName)
	}

	if report.StartDate != "01.01.2026" {
		t.Errorf("expected StartDate '01.01.2026', got '%s'", report.StartDate)
	}

	if report.EndDate != "31.01.2026" {
		t.Errorf("expected EndDate '31.01.2026', got '%s'", report.EndDate)
	}

	if report.SerialNumber != "ABC123" {
		t.Errorf("expected SerialNumber 'ABC123', got '%s'", report.SerialNumber)
	}

	if report.LicensePlate != "B-GO 123" {
		t.Errorf("expected LicensePlate 'B-GO 123', got '%s'", report.LicensePlate)
	}

	if report.Mileage != "50000 km" {
		t.Errorf("expected Mileage '50000 km', got '%s'", report.Mileage)
	}

	if report.KwhPrice != 0.35 {
		t.Errorf("expected KwhPrice 0.35, got %.2f", report.KwhPrice)
	}

	if report.TotalSessions != 5 {
		t.Errorf("expected TotalSessions 5, got %d", report.TotalSessions)
	}

	if report.TotalEnergy != 50.0 {
		t.Errorf("expected TotalEnergy 50.0, got %.2f", report.TotalEnergy)
	}

	if report.TotalPrice != 17.50 {
		t.Errorf("expected TotalPrice 17.50, got %.2f", report.TotalPrice)
	}

	if len(report.Sessions) != 1 {
		t.Errorf("expected 1 session, got %d", len(report.Sessions))
	}
}

func TestReportData_EmptySessions(t *testing.T) {
	report := ReportData{
		MonthName:     "01-2026",
		TotalSessions: 0,
		Sessions:      nil,
	}

	if report.TotalSessions != 0 {
		t.Errorf("expected TotalSessions 0, got %d", report.TotalSessions)
	}

	if report.Sessions != nil {
		t.Errorf("expected nil Sessions, got %v", report.Sessions)
	}
}
