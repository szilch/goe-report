package cmd

import (
	"echarge-report/pkg/config"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestStatusCmd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"car": 1, "alw": true, "amp": 16, "wh": 500.0, "eto": 1000.0, "nrg": [230, 231, 229, 0, 10, 11, 9, 2300, 2541, 2061, 0, 6902], "tma": [30.0], "frc": 0}`)
	}))
	defer server.Close()

	defer viper.Reset()
	viper.Set(config.KeyWallboxGoeLocalApiUrl, server.URL)

	output, err := executeCommand(rootCmd, "status")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(output, "Wallbox Status Report:") {
		t.Errorf("Expected output to contain report header, got: %s", output)
	}
	if !strings.Contains(output, "Idle (not connected)") {
		t.Errorf("Expected vehicle state 'Idle (not connected)', got: %s", output)
	}
}
