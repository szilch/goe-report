package formatter

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestNewTerminalFormatter(t *testing.T) {
	f := NewTerminalFormatter()

	if f == nil {
		t.Error("NewTerminalFormatter() returned nil")
	}
}

func TestTerminalFormatter_Format_Output(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f := NewTerminalFormatter()
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

	// Restore stdout
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check for expected content
	if !strings.Contains(output, "Ladehistorie für Wallbox") {
		t.Error("output should contain 'Ladehistorie für Wallbox'")
	}

	if !strings.Contains(output, "B-GO 123") {
		t.Error("output should contain license plate 'B-GO 123'")
	}

	if !strings.Contains(output, "50000 km") {
		t.Error("output should contain mileage '50000 km'")
	}

	if !strings.Contains(output, "01.01.2026 - 31.01.2026") {
		t.Error("output should contain date range")
	}

	if !strings.Contains(output, "Gesamte Ladevorgänge") {
		t.Error("output should contain 'Gesamte Ladevorgänge'")
	}

	if !strings.Contains(output, "20.00 kWh") {
		t.Error("output should contain total energy '20.00 kWh'")
	}
}

func TestTerminalFormatter_Format_NoSessions(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f := NewTerminalFormatter()
	data := ReportData{
		MonthName:     "01-2026",
		StartDate:     "01.01.2026",
		EndDate:       "31.01.2026",
		TotalSessions: 0,
		Sessions:      nil,
	}

	err := f.Format(data)

	// Restore stdout
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Keine Ladevorgänge") {
		t.Error("output should contain 'Keine Ladevorgänge' when no sessions")
	}
}

func TestTerminalFormatter_Format_NoLicensePlate(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f := NewTerminalFormatter()
	data := ReportData{
		MonthName:     "01-2026",
		StartDate:     "01.01.2026",
		EndDate:       "31.01.2026",
		LicensePlate:  "", // Empty license plate
		TotalSessions: 0,
	}

	err := f.Format(data)

	// Restore stdout
	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "Keines hinterlegt") {
		t.Error("output should contain 'Keines hinterlegt' when no license plate")
	}
}

func TestTerminalFormatter_ImplementsFormatter(t *testing.T) {
	var _ Formatter = &TerminalFormatter{}
}
