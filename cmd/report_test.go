package cmd

import (
	"testing"
)

func TestReportCmd_Flags(t *testing.T) {
	// Test if it's there
	if reportCmd.Use != "report" {
		t.Errorf("Expected command use 'report', got: %s", reportCmd.Use)
	}

	// Testing flags initialization
	if reportCmd.Flags().Lookup("month") == nil {
		t.Error("Flag --month should be defined")
	}
	if reportCmd.Flags().Lookup("pdf") == nil {
		t.Error("Flag --pdf should be defined")
	}
}

// We don't test the Run part here because it calls os.Exit(1) which would kill the test.
// In a real application, we would refactor the Run logic into a separate function that returns an error.
