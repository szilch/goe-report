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

		// History API returns array of arrays
		history := [][]stateResponse{
			{
				{State: "50000"},
			},
		}
		data, _ := json.Marshal(history)
		rw.Write(data)
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "test-token", "sensor.test_mileage")

	value, err := p.GetMileage()
	if err != nil {
		t.Fatalf("GetMileage failed: %v", err)
	}

	if value != 50000 {
		t.Errorf("Expected 50000, got %d", value)
	}
}

func TestHomeAssistantProvider_GetMileageAt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Bearer test-token, got %s", req.Header.Get("Authorization"))
		}

		// Check if URL contains history path
		if !strings.Contains(req.URL.Path, "/api/history/period/") {
			t.Errorf("Expected history API path, got %s", req.URL.Path)
		}

		history := [][]stateResponse{
			{
				{State: "48500"},
			},
		}
		data, _ := json.Marshal(history)
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
		rw.Write([]byte(`[]`))
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "test-token", "sensor.test")

	targetTime := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)
	_, err := p.GetMileageAt(targetTime)
	if err == nil {
		t.Fatalf("Expected error for empty history, got nil")
	}
}

func TestHomeAssistantProvider_GetMileageAt_EmptyInnerArray(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`[[]]`))
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "test-token", "sensor.test")

	targetTime := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)
	_, err := p.GetMileageAt(targetTime)
	if err == nil {
		t.Fatalf("Expected error for empty inner array, got nil")
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
