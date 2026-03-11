package config

import (
	"testing"
)

func TestConfigDirName(t *testing.T) {
	expected := ".goe-report"
	if ConfigDirName != expected {
		t.Errorf("expected ConfigDirName '%s', got '%s'", expected, ConfigDirName)
	}
}

func TestConfigFileName(t *testing.T) {
	expected := ".goereportrc"
	if ConfigFileName != expected {
		t.Errorf("expected ConfigFileName '%s', got '%s'", expected, ConfigFileName)
	}
}

func TestKeyConstants(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"KeyToken", KeyToken, "goe_token"},
		{"KeyLocalApiUrl", KeyLocalApiUrl, "goe_localApiUrl"},
		{"KeySerial", KeySerial, "goe_serial"},
		{"KeyChipIds", KeyChipIds, "goe_chipIds"},
		{"KeyLicensePlate", KeyLicensePlate, "licenseplate"},
		{"KeyKwhPrice", KeyKwhPrice, "kwhprice"},
		{"KeyHAToken", KeyHAToken, "ha_token"},
		{"KeyHAAPI", KeyHAAPI, "ha_api"},
		{"KeyHAMilageSensor", KeyHAMilageSensor, "ha_milage_sensorid"},
		{"KeyMailHost", KeyMailHost, "mail_host"},
		{"KeyMailPort", KeyMailPort, "mail_port"},
		{"KeyMailUsername", KeyMailUsername, "mail_username"},
		{"KeyMailPassword", KeyMailPassword, "mail_password"},
		{"KeyMailFrom", KeyMailFrom, "mail_from"},
		{"KeyMailTo", KeyMailTo, "mail_to"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.key != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, tt.key)
			}
		})
	}
}

func TestKeyConstants_NotEmpty(t *testing.T) {
	keys := []string{
		KeyToken,
		KeyLocalApiUrl,
		KeySerial,
		KeyChipIds,
		KeyLicensePlate,
		KeyKwhPrice,
		KeyHAToken,
		KeyHAAPI,
		KeyHAMilageSensor,
		KeyMailHost,
		KeyMailPort,
		KeyMailUsername,
		KeyMailPassword,
		KeyMailFrom,
		KeyMailTo,
	}

	for _, key := range keys {
		if key == "" {
			t.Errorf("key should not be empty")
		}
	}
}
