package pdf

import (
	"echarge-report/pkg/config"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

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
			continue
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

func (s *Service) merge(dst string, srcs []string) error {
	if len(srcs) == 0 {
		return nil
	}

	allFiles := append([]string{dst}, srcs...)

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

	if err := os.Rename(tmpPath, dst); err != nil {
		return fmt.Errorf("failed to replace report PDF with merged file: %w", err)
	}
	return nil
}
