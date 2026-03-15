package pdf

import (
	"echarge-report/pkg/config"
	"os"
	"path/filepath"
	"testing"

	"github.com/jung-kurt/gofpdf"
)

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

	basePDF := filepath.Join(tempDir, "base.pdf")
	createDummyPDF(t, basePDF, "Base PDF")

	attach1 := filepath.Join(tempDir, "attach1.pdf")
	createDummyPDF(t, attach1, "Attach 1")

	attach2 := filepath.Join(tempDir, "attach2.pdf")
	createDummyPDF(t, attach2, "Attach 2")

	s := NewService()

	err := s.merge(basePDF, []string{attach1, attach2})
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	info, err := os.Stat(basePDF)
	if err != nil {
		t.Fatalf("base.pdf disappeared after merge: %v", err)
	}
	if info.Size() == 0 {
		t.Errorf("merged PDF is empty")
	}

	err = s.merge(basePDF, []string{})
	if err != nil {
		t.Fatalf("merge failed with empty slice: %v", err)
	}
}

func TestService_AttachExistingPDFsToReport(t *testing.T) {
	tempHome := t.TempDir()

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tempHome, config.ConfigDirName)
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	attach1 := filepath.Join(configDir, "attach1.pdf")
	createDummyPDF(t, attach1, "Attach 1")

	attach2 := filepath.Join(configDir, "attach2.pdf")
	createDummyPDF(t, attach2, "Attach 2")

	os.WriteFile(filepath.Join(configDir, "not_a_pdf.txt"), []byte("ignore me"), 0644)

	reportPDF := filepath.Join(tempHome, "report.pdf")
	createDummyPDF(t, reportPDF, "Report PDF")

	s := NewService()

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

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tempHome, config.ConfigDirName)
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	attach1 := filepath.Join(configDir, "attach1.pdf")
	createDummyPDF(t, attach1, "Attach 1")

	reportPDF := filepath.Join(configDir, "report.pdf")
	createDummyPDF(t, reportPDF, "Report PDF")

	s := NewService()

	count, _, err := s.AttachExistingPDFsToReport(reportPDF)
	if err != nil {
		t.Fatalf("AttachExistingPDFsToReport failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 PDF to be attached (skipping self), got %d", count)
	}
}

func TestService_AttachExistingPDFsToReport_NoAttachments(t *testing.T) {
	tempHome := t.TempDir()

	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", origHome)

	configDir := filepath.Join(tempHome, config.ConfigDirName)
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

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
