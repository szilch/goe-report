package formatter

import (
	"echarge-report/pkg/models"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPDFFormatter_Format(t *testing.T) {
	testDir := t.TempDir()
	testFile := filepath.Join(testDir, "test_report.pdf")

	formatter := NewPDFFormatter(testFile)

	data := models.ReportData{
		LicensePlate:  "W-TEST123",
		Mileage:      12345,
		MileageAtEnd: 12500,
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
	formatter := NewPDFFormatter("/nonexistent_dir_12345/test_report.pdf")

	data := models.ReportData{}

	err := formatter.Format(data)
	if err == nil {
		t.Fatalf("Expected error when saving to invalid path, got nil")
	}
}
