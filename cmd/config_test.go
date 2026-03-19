package cmd

import (
	"echarge-report/pkg/config"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestIsAllowedKey(t *testing.T) {
	tests := []struct {
		key     string
		allowed bool
	}{
		{config.KeyWallboxGoeCloudSerial, true},
		{config.KeyLicensePlate, true},
		{"unknown-key", false},
	}

	for _, tt := range tests {
		_, ok := isAllowedKey(tt.key)
		if ok != tt.allowed {
			t.Errorf("isAllowedKey(%s) = %v, expected %v", tt.key, ok, tt.allowed)
		}
	}
}

func TestKeyList(t *testing.T) {
	list := keyList()
	if !strings.Contains(list, config.KeyWallboxGoeCloudSerial) {
		t.Errorf("keyList() output does not contain expected key: %s", config.KeyWallboxGoeCloudSerial)
	}
}

func TestConfigGetCmd(t *testing.T) {
	defer viper.Reset()
	viper.Set(config.KeyLicensePlate, "ABC-123")

	output, err := executeCommand(rootCmd, "config-get", config.KeyLicensePlate)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(output, "ABC-123") {
		t.Errorf("Expected output to contain 'ABC-123', got: %s", output)
	}
}

func TestConfigListCmd(t *testing.T) {
	defer viper.Reset()
	viper.Set(config.KeyLicensePlate, "XYZ-789")

	output, err := executeCommand(rootCmd, "config-list")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(output, "Current configuration:") {
		t.Error("Expected output to contain header")
	}
	if !strings.Contains(output, "XYZ-789") {
		t.Errorf("Expected output to contain 'XYZ-789', got: %s", output)
	}
}
