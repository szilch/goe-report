package pdfmerge

import (
	"os"
	"path/filepath"
	"testing"
)

func createTestPDF(t *testing.T, filename string) {
	t.Helper()
	// Create a minimal valid PDF file
	content := `%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>
endobj
xref
0 4
0000000000 65535 f
0000000009 00000 n
0000000058 00000 n
0000000115 00000 n
trailer
<< /Root 1 0 R /Size 4 >>
startxref
191
%%EOF`
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create test PDF: %v", err)
	}
}

func TestMerge_EmptySources(t *testing.T) {
	err := Merge("dst.pdf", []string{})

	if err != nil {
		t.Errorf("expected no error for empty sources, got: %v", err)
	}
}

func TestMerge_NilSources(t *testing.T) {
	err := Merge("dst.pdf", nil)

	if err != nil {
		t.Errorf("expected no error for nil sources, got: %v", err)
	}
}

func TestMerge_Success(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "pdfmerge_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test PDFs
	dst := filepath.Join(tmpDir, "main.pdf")
	src1 := filepath.Join(tmpDir, "attach1.pdf")
	src2 := filepath.Join(tmpDir, "attach2.pdf")

	createTestPDF(t, dst)
	createTestPDF(t, src1)
	createTestPDF(t, src2)

	// Merge
	err = Merge(dst, []string{src1, src2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that destination file still exists
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		t.Error("merged PDF was not created")
	}

	// Check file size increased (merged file should be larger)
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("failed to stat merged file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("merged PDF is empty")
	}
}

func TestMerge_NonExistentSource(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "pdfmerge_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dst := filepath.Join(tmpDir, "main.pdf")
	createTestPDF(t, dst)

	// Try to merge with non-existent source
	err = Merge(dst, []string{filepath.Join(tmpDir, "nonexistent.pdf")})

	if err == nil {
		t.Error("expected error for non-existent source, got nil")
	}
}

func TestMerge_NonExistentDestination(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "pdfmerge_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	src := filepath.Join(tmpDir, "source.pdf")
	createTestPDF(t, src)

	dst := filepath.Join(tmpDir, "nonexistent.pdf")

	// Try to merge with non-existent destination
	err = Merge(dst, []string{src})

	if err == nil {
		t.Error("expected error for non-existent destination, got nil")
	}
}

func TestMerge_SingleSource(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "pdfmerge_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dst := filepath.Join(tmpDir, "main.pdf")
	src := filepath.Join(tmpDir, "attach.pdf")

	createTestPDF(t, dst)
	createTestPDF(t, src)

	// Merge single source
	err = Merge(dst, []string{src})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that destination file exists
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		t.Error("merged PDF was not created")
	}
}

func TestMerge_InvalidPDF(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "pdfmerge_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dst := filepath.Join(tmpDir, "main.pdf")
	src := filepath.Join(tmpDir, "invalid.pdf")

	createTestPDF(t, dst)

	// Create invalid PDF
	err = os.WriteFile(src, []byte("not a valid pdf"), 0644)
	if err != nil {
		t.Fatalf("failed to create invalid PDF: %v", err)
	}

	// Try to merge with invalid source
	err = Merge(dst, []string{src})

	if err == nil {
		t.Error("expected error for invalid PDF, got nil")
	}
}
