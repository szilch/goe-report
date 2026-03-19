package pdf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// Service provides capabilities for manipulating PDF files.
type Service struct{}

// NewService creates a new Service for PDF operations.
func NewService() *Service {
	return &Service{}
}

// AttachExistingPDFsToReport merges any .pdf files found in the specified
// attachDir into the end of the reportFile. It returns the number of
// files attached and the directory that was scanned.
func (s *Service) AttachExistingPDFsToReport(reportFile string, attachDir string) (int, string, error) {
	matches, err := filepath.Glob(filepath.Join(attachDir, "*.pdf"))
	if err != nil {
		return 0, "", fmt.Errorf("scan for attachment PDFs: %w", err)
	}

	reportAbs, _ := filepath.Abs(reportFile)

	var attachments []string
	for _, m := range matches {
		abs, _ := filepath.Abs(m)
		if abs == reportAbs {
			continue
		}
		attachments = append(attachments, m)
	}

	if len(attachments) == 0 {
		return 0, attachDir, nil
	}

	if err := s.merge(reportFile, attachments); err != nil {
		return 0, attachDir, fmt.Errorf("attach PDFs: %w", err)
	}
	return len(attachments), attachDir, nil
}

func (s *Service) merge(dst string, srcs []string) error {
	if len(srcs) == 0 {
		return nil
	}

	allFiles := append([]string{dst}, srcs...)

	tmpFile, err := os.CreateTemp("", "echarge-report-merge-*.pdf")
	if err != nil {
		return fmt.Errorf("create temp file for merge: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := api.MergeCreateFile(allFiles, tmpPath, false, nil); err != nil {
		return fmt.Errorf("merge PDFs: %w", err)
	}

	if err := os.Rename(tmpPath, dst); err != nil {
		return fmt.Errorf("replace report PDF with merged file: %w", err)
	}
	return nil
}
