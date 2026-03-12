package formatter

import (
	"echarge-report/pkg/models"
	"os"
	"path/filepath"
	"testing"
)

func TestPDFFormatter_Format(t *testing.T) {
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test_report.pdf")

	formatter := NewPDFFormatter(testFile)

	data := models.ReportData{
		LicensePlate:  "W-TEST123",
		Mileage:       "12345",
		StartDate:     "01.01.2026",
		EndDate:       "31.01.2026",
		KwhPrice:      0.30,
		TotalSessions: 2,
		TotalEnergy:   50.0,
		TotalPrice:    15.0,
		Sessions: []models.SessionData{
			{
				StartDate: "01.01.2026 10:00",
				EndDate:   "01.01.2026 12:00",
				Duration:  "2h0m",
				Energy:    20.0,
				Price:     6.0,
			},
			{
				StartDate: "02.01.2026 14:00",
				EndDate:   "02.01.2026 17:00",
				Duration:  "3h0m",
				Energy:    30.0,
				Price:     9.0,
			},
		},
	}

	err := formatter.Format(data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	info, err := os.Stat(testFile)
	if os.IsNotExist(err) {
		t.Fatalf("Expected PDF file to be created, but it was not")
	}
	if info.Size() == 0 {
		t.Errorf("Expected PDF file to have size > 0, got 0")
	}
}

func TestPDFFormatter_Format_NoSessions(t *testing.T) {
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test_report_empty.pdf")

	formatter := NewPDFFormatter(testFile)

	data := models.ReportData{
		TotalSessions: 0,
	}

	err := formatter.Format(data)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	info, err := os.Stat(testFile)
	if os.IsNotExist(err) {
		t.Fatalf("Expected PDF file to be created, but it was not")
	}
	if info.Size() == 0 {
		t.Errorf("Expected PDF file to have size > 0, got 0")
	}
}

func TestPDFFormatter_Format_InvalidPath(t *testing.T) {
	// Trying to write to a path inside a directory that does not exist
	formatter := NewPDFFormatter("/nonexistent_dir_12345/test_report.pdf")

	data := models.ReportData{}

	err := formatter.Format(data)
	if err == nil {
		t.Fatalf("Expected error when saving to invalid path, got nil")
	}
}
