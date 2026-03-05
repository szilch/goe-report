package pdfmerge

import (
	"fmt"
	"os"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// Merge appends the pages from each PDF in srcs (in order) to the dst PDF file.
// The dst file is modified in-place and will contain its original content followed
// by the pages of all srcs.
func Merge(dst string, srcs []string) error {
	if len(srcs) == 0 {
		return nil
	}

	// pdfcpu's MergeAppendFile merges srcs into dst in-place.
	// We build the full list: dst first, then all attachments.
	allFiles := append([]string{dst}, srcs...)

	// Write the merged result to a temporary file first, then replace dst.
	tmpFile, err := os.CreateTemp("", "goe-report-merge-*.pdf")
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
