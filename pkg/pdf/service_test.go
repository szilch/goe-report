package pdf

import (
	"echarge-report/pkg/config"
	"os"
	"path/filepath"
	"testing"

	"github.com/jung-kurt/gofpdf"
)

// createDummyPDF creates a minimal valid PDF file for testing
func createDummyPDF(t *testing.T, filename string, text string) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, text)
	err := pdf.OutputFileAndClose(filename)
	if err != nil {
		t.Fatalf("Failed to create dummy PDF %s: %v", filename, err)
	}
}

func TestService_merge(t *testing.T) {
	tempDir := t.TempDir()

	// Create base PDF
	basePDF := filepath.Join(tempDir, "base.pdf")
	createDummyPDF(t, basePDF, "Base PDF")

	// Create attachment PDFs
	attach1 := filepath.Join(tempDir, "attach1.pdf")
	createDummyPDF(t, attach1, "Attach 1")

	attach2 := filepath.Join(tempDir, "attach2.pdf")
	createDummyPDF(t, attach2, "Attach 2")

	s := NewService()

	// Test normal merge
	err := s.merge(basePDF, []string{attach1, attach2})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// Verify basePDF exists and has size
	info, err := os.Stat(basePDF)
	if err != nil {
		t.Fatalf("base.pdf disappeared after merge: %v", err)
	}
	if info.Size() == 0 {
		t.Errorf("merged PDF is empty")
	}

	// Test empty slice merge
	err = s.merge(basePDF, []string{})
	if err != nil {
		t.Fatalf("merge failed with empty slice: %v", err)
	}
}

func TestService_AttachExistingPDFsToReport(t *testing.T) {
	tempHome := t.TempDir()

	// Override HOME directory for os.UserHomeDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	// Create config directory
	configDir := filepath.Join(tempHome, config.ConfigDirName)
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Create some dummy PDFs in config dir
	attach1 := filepath.Join(configDir, "attach1.pdf")
	createDummyPDF(t, attach1, "Attach 1")

	attach2 := filepath.Join(configDir, "attach2.pdf")
	createDummyPDF(t, attach2, "Attach 2")

	// Also create a non-PDF file to ensure it's ignored
	os.WriteFile(filepath.Join(configDir, "not_a_pdf.txt"), []byte("ignore me"), 0644)

	// Create the report PDF somewhere else
	reportPDF := filepath.Join(tempHome, "report.pdf")
	createDummyPDF(t, reportPDF, "Report PDF")

	s := NewService()

	// Test attachment
	count, usedConfigDir, err := s.AttachExistingPDFsToReport(reportPDF)
	if err != nil {
		t.Fatalf("AttachExistingPDFsToReport failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 PDFs to be attached, got %d", count)
	}

	if usedConfigDir != configDir {
		t.Errorf("expected config dir %s, got %s", configDir, usedConfigDir)
	}

	// Verify report PDF still exists
	info, err := os.Stat(reportPDF)
	if os.IsNotExist(err) {
		t.Fatalf("report PDF missing after attachment")
	}
	if info.Size() == 0 {
		t.Errorf("report PDF is empty after attachment")
	}
}

func TestService_AttachExistingPDFsToReport_SkipSelf(t *testing.T) {
	tempHome := t.TempDir()

	// Override HOME directory for os.UserHomeDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	// Create config directory
	configDir := filepath.Join(tempHome, config.ConfigDirName)
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Create attachment PDF
	attach1 := filepath.Join(configDir, "attach1.pdf")
	createDummyPDF(t, attach1, "Attach 1")

	// Create the report PDF IN the config dir, so we verify it's skipped
	reportPDF := filepath.Join(configDir, "report.pdf")
	createDummyPDF(t, reportPDF, "Report PDF")

	s := NewService()

	count, _, err := s.AttachExistingPDFsToReport(reportPDF)
	if err != nil {
		t.Fatalf("AttachExistingPDFsToReport failed: %v", err)
	}

	// It should skip report.pdf and only attach attach1.pdf
	if count != 1 {
		t.Errorf("expected 1 PDF to be attached (skipping self), got %d", count)
	}
}

func TestService_AttachExistingPDFsToReport_NoAttachments(t *testing.T) {
	tempHome := t.TempDir()

	// Override HOME directory for os.UserHomeDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	// Create config directory
	configDir := filepath.Join(tempHome, config.ConfigDirName)
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Config dir exists but has no PDFs

	// Create the report PDF
	reportPDF := filepath.Join(tempHome, "report.pdf")
	createDummyPDF(t, reportPDF, "Report PDF")

	s := NewService()

	count, _, err := s.AttachExistingPDFsToReport(reportPDF)
	if err != nil {
		t.Fatalf("AttachExistingPDFsToReport failed: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 PDFs to be attached, got %d", count)
	}
}
