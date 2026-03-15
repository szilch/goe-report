package formatter

import (
	"bytes"
	"echarge-report/pkg/models"
	"os"
	"strings"
	"testing"
	"time"
)

// captureStdout captures the output printed to os.Stdout during the execution of f.
func captureStdout(f func() error) (string, error) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := f()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String(), err
}

func TestTerminalFormatter_Format(t *testing.T) {
	formatter := NewTerminalFormatter()

	data := models.ReportData{
		LicensePlate:  "W-TEST123",
		Mileage:       "12345",
		StartDate:     time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:       time.Date(2026, 1, 31, 23, 59, 59, 999, time.UTC),
		KwhPrice:      0.30,
		TotalSessions: 2,
		TotalEnergy:   50.0,
		TotalPrice:    15.0,
		Sessions: []models.SessionData{
			{
				StartDate: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC),
				Duration:  "2h0m",
				Energy:    20.0,
				Price:     6.0,
			},
			{
				StartDate: time.Date(2026, 1, 2, 14, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2026, 1, 2, 17, 0, 0, 0, time.UTC),
				Duration:  "3h0m",
				Energy:    30.0,
				Price:     9.0,
			},
		},
	}

	output, err := captureStdout(func() error {
		return formatter.Format(data)
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedStrings := []string{
		"Ladehistorie",
		"Kfz-Kennzeichen:",
		"W-TEST123",
		"01.01.2026 - 31.01.2026",
		"0,3000 €",
		"20.00 kWh",
		"6,00 €",
		"30.00 kWh",
		"9,00 €",
		"2",
		"50.00 kWh",
		"15,00 €",
	}

	for _, s := range expectedStrings {
		if !strings.Contains(output, s) {
			t.Errorf("Expected output to contain %q, but it did not", s)
		}
	}
}

func TestTerminalFormatter_Format_NoSessions(t *testing.T) {
	formatter := NewTerminalFormatter()

	data := models.ReportData{
		TotalSessions: 0,
	}

	output, err := captureStdout(func() error {
		return formatter.Format(data)
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedString := "Keine Ladevorgänge für diese Kriterien im gewünschten Zeitraum gefunden."
	if !strings.Contains(output, expectedString) {
		t.Errorf("Expected output to contain %q, but it did not", expectedString)
	}
}

func TestTerminalFormatter_Format_NoLicensePlate(t *testing.T) {
	formatter := NewTerminalFormatter()

	data := models.ReportData{
		LicensePlate:  "",
		TotalSessions: 0,
	}

	output, err := captureStdout(func() error {
		return formatter.Format(data)
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	expectedString := "Keines hinterlegt"
	if !strings.Contains(output, expectedString) {
		t.Errorf("Expected output to contain %q, but it did not", expectedString)
	}
}
