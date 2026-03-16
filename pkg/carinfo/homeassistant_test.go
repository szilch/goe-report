package carinfo

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHomeAssistantProvider_GetType(t *testing.T) {
	p := &HomeAssistantProvider{}
	if p.GetType() != "homeassistant" {
		t.Errorf("Expected 'homeassistant', got '%s'", p.GetType())
	}
}

func newTestProvider(serverURL, token, sensorID string) *HomeAssistantProvider {
	return &HomeAssistantProvider{
		apiURL:   serverURL,
		token:    token,
		sensorID: sensorID,
		client:   &http.Client{},
	}
}

func TestHomeAssistantProvider_GetMileage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Bearer test-token, got %s", req.Header.Get("Authorization"))
		}

		// Statistics API
		result := map[string][]struct {
			Mean  *float64 `json:"mean"`
			State *float64 `json:"state"`
		}{
			"sensor.test_mileage": {
				{Mean: floatPtr(50000.7)},
			},
		}
		data, _ := json.Marshal(result)
		rw.Write(data)
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "test-token", "sensor.test_mileage")

	value, err := p.GetMileage()
	if err != nil {
		t.Fatalf("GetMileage failed: %v", err)
	}

	if value != 50000 {
		t.Errorf("Expected 50000 (truncated), got %d", value)
	}
}

func TestHomeAssistantProvider_GetMileageAt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Bearer test-token, got %s", req.Header.Get("Authorization"))
		}

		// Check if URL contains statistics path
		if !strings.Contains(req.URL.Path, "/api/statistics/during_period") {
			t.Errorf("Expected statistics API path, got %s", req.URL.Path)
		}

		result := map[string][]struct {
			Mean  *float64 `json:"mean"`
			State *float64 `json:"state"`
		}{
			"sensor.test_mileage": {
				{State: floatPtr(48500.0)},
			},
		}
		data, _ := json.Marshal(result)
		rw.Write(data)
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "test-token", "sensor.test_mileage")

	targetTime := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)
	value, err := p.GetMileageAt(targetTime)
	if err != nil {
		t.Fatalf("GetMileageAt failed: %v", err)
	}

	if value != 48500 {
		t.Errorf("Expected 48500, got %d", value)
	}
}

func TestHomeAssistantProvider_GetMileageAt_NoData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`{}`))
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "test-token", "sensor.test")

	targetTime := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)
	_, err := p.GetMileageAt(targetTime)
	if err == nil {
		t.Fatalf("Expected error for empty history, got nil")
	}
	if !strings.Contains(err.Error(), "keine Langzeitdaten") {
		t.Errorf("Expected error to mention 'keine Langzeitdaten', got: %v", err)
	}
}

func TestHomeAssistantProvider_GetMileage_MissingConfig(t *testing.T) {
	p := &HomeAssistantProvider{apiURL: "", token: "tok", sensorID: "id", client: &http.Client{}}
	_, err := p.GetMileage()
	if err == nil {
		t.Errorf("Expected error for missing apiURL, got nil")
	}

	p = &HomeAssistantProvider{apiURL: "http://localhost", token: "", sensorID: "id", client: &http.Client{}}
	_, err = p.GetMileage()
	if err == nil {
		t.Errorf("Expected error for missing token, got nil")
	}

	p = &HomeAssistantProvider{apiURL: "http://localhost", token: "tok", sensorID: "", client: &http.Client{}}
	_, err = p.GetMileage()
	if err == nil {
		t.Errorf("Expected error for missing sensorID, got nil")
	}
}

func TestHomeAssistantProvider_GetMileage_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`this is not json`))
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "test-token", "sensor.test")

	_, err := p.GetMileage()
	if err == nil {
		t.Fatalf("Expected error for invalid JSON, got nil")
	}
}

func floatPtr(f float64) *float64 {
	return &f
}
