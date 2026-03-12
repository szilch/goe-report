package pdf

import (
	"echarge-report/pkg/config"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// Service provides PDF manipulation capabilities.
type Service struct{}

// NewService creates a new PDF service.
func NewService() *Service {
	return &Service{}
}

// AttachExistingPDFsToReport finds all PDFs in the configuration directory and attaches them to the report PDF.
// It returns the number of attached files and the configuration directory used, or an error.
func (s *Service) AttachExistingPDFsToReport(reportFile string) (int, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return 0, "", fmt.Errorf("error determining home directory: %w", err)
	}
	configDir := filepath.Join(home, config.ConfigDirName)

	matches, err := filepath.Glob(filepath.Join(configDir, "*.pdf"))
	if err != nil {
		return 0, "", fmt.Errorf("error scanning for attachment PDFs: %w", err)
	}

	reportAbs, _ := filepath.Abs(reportFile)

	var attachments []string
	for _, m := range matches {
		abs, _ := filepath.Abs(m)
		if abs == reportAbs {
			continue // skip the report itself
		}
		attachments = append(attachments, m)
	}

	if len(attachments) == 0 {
		return 0, configDir, nil
	}

	if err := s.merge(reportFile, attachments); err != nil {
		return 0, configDir, fmt.Errorf("error attaching PDFs: %w", err)
	}
	return len(attachments), configDir, nil
}

// merge appends the pages from each PDF in srcs (in order) to the dst PDF file.
// The dst file is modified in-place and will contain its original content followed
// by the pages of all srcs.
func (s *Service) merge(dst string, srcs []string) error {
	if len(srcs) == 0 {
		return nil
	}

	// pdfcpu's MergeAppendFile merges srcs into dst in-place.
	// We build the full list: dst first, then all attachments.
	allFiles := append([]string{dst}, srcs...)

	// Write the merged result to a temporary file first, then replace dst.
	tmpFile, err := os.CreateTemp("", "echarge-report-merge-*.pdf")
	if err != nil {
		return fmt.Errorf("failed to create temp file for merge: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	if err := api.MergeCreateFile(allFiles, tmpPath, false, nil); err != nil {
		return fmt.Errorf("failed to merge PDFs: %w", err)
	}

	// Replace dst with the merged result.
	if err := os.Rename(tmpPath, dst); err != nil {
		return fmt.Errorf("failed to replace report PDF with merged file: %w", err)
	}
	return nil
}
