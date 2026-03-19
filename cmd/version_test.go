package cmd

import (
	"bytes"
	"echarge-report/pkg/version"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func executeCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err = root.Execute()
	return buf.String(), err
}

func TestVersionCmd(t *testing.T) {
	version.Version = "v1.2.3-test"
	output, err := executeCommand(rootCmd, "version")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := "echarge-report v1.2.3-test"
	if !strings.Contains(output, expected) {
		t.Errorf("Expected output to contain '%s', got: %s", expected, output)
	}
}
