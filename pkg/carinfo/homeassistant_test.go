package carinfo

import (
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
	// GetMileage now calls GetMileageAt(time.Now()), so we need to mock the History API
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Bearer test-token, got %s", req.Header.Get("Authorization"))
		}

		// History API returns array of arrays
		rw.Write([]byte(`[[
			{"state": "50000", "attributes": {"unit_of_measurement": "km"}, "last_changed": "2026-03-16T10:00:00Z"}
		]]`))
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "test-token", "sensor.test_mileage")

	value, err := p.GetMileage()
	if err != nil {
		t.Fatalf("GetMileage failed: %v", err)
	}

	if value != "50000 km" {
		t.Errorf("Expected '50000 km', got '%s'", value)
	}
}

func TestHomeAssistantProvider_GetMileage_NoUnit(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestCount++

		// First request is history API (no unit)
		if requestCount == 1 {
			rw.Write([]byte(`[[{"state": "42", "attributes": {}}]]`))
			return
		}

		// Second request is current state (to get unit) - return no unit
		rw.Write([]byte(`{"state": "42", "attributes": {}}`))
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "test-token", "sensor.test")

	value, err := p.GetMileage()
	if err != nil {
		t.Fatalf("GetMileage failed: %v", err)
	}

	if value != "42" {
		t.Errorf("Expected '42', got '%s'", value)
	}
}

func TestHomeAssistantProvider_GetMileage_UnavailableState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// History API format
		rw.Write([]byte(`[[{"state": "unavailable", "attributes": {}}]]`))
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "test-token", "sensor.test")

	_, err := p.GetMileage()
	if err == nil {
		t.Fatalf("Expected error for unavailable state, got nil")
	}
}

func TestHomeAssistantProvider_GetMileage_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write([]byte(`{"message": "401: Unauthorized"}`))
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "wrong-token", "sensor.test")

	_, err := p.GetMileage()
	if err == nil {
		t.Fatalf("Expected error for HTTP 401, got nil")
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

// Tests for GetMileageAt (History API)

func TestHomeAssistantProvider_GetMileageAt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Bearer test-token, got %s", req.Header.Get("Authorization"))
		}

		// History API returns array of arrays with multiple entries
		// The code should pick the LAST entry (most recent before target time)
		rw.Write([]byte(`[[
			{"state": "48000", "attributes": {"unit_of_measurement": "km"}, "last_changed": "2026-01-30T10:00:00Z"},
			{"state": "48500", "attributes": {"unit_of_measurement": "km"}, "last_changed": "2026-01-31T18:00:00Z"}
		]]`))
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "test-token", "sensor.test_mileage")

	targetTime := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)
	value, err := p.GetMileageAt(targetTime)
	if err != nil {
		t.Fatalf("GetMileageAt failed: %v", err)
	}

	// Should return the LAST entry (48500), not the first (48000)
	if value != "48500 km" {
		t.Errorf("Expected '48500 km', got '%s'", value)
	}
}

func TestHomeAssistantProvider_GetMileageAt_NoUnit_FallbackToCurrentState(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestCount++

		// First request is history API (no unit in response)
		if requestCount == 1 {
			rw.Write([]byte(`[[
				{"state": "48000", "attributes": {}},
				{"state": "48500", "attributes": {}}
			]]`))
			return
		}

		// Second request is current state (to get unit)
		rw.Write([]byte(`{"state": "49000", "attributes": {"unit_of_measurement": "km"}}`))
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "test-token", "sensor.test_mileage")

	targetTime := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)
	value, err := p.GetMileageAt(targetTime)
	if err != nil {
		t.Fatalf("GetMileageAt failed: %v", err)
	}

	// Should return the LAST history entry (48500) with unit from current state
	if value != "48500 km" {
		t.Errorf("Expected '48500 km', got '%s'", value)
	}
}

func TestHomeAssistantProvider_GetMileageAt_NoHistoryData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Empty history response
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
		// History with empty inner array
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

func TestHomeAssistantProvider_GetMileageAt_UnavailableState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`[[{"state": "unavailable", "attributes": {}}]]`))
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "test-token", "sensor.test")

	targetTime := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)
	_, err := p.GetMileageAt(targetTime)
	if err == nil {
		t.Fatalf("Expected error for unavailable state, got nil")
	}
}

func TestHomeAssistantProvider_GetMileageAt_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write([]byte(`{"message": "401: Unauthorized"}`))
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "wrong-token", "sensor.test")

	targetTime := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)
	_, err := p.GetMileageAt(targetTime)
	if err == nil {
		t.Fatalf("Expected error for HTTP 401, got nil")
	}
}

func TestHomeAssistantProvider_GetMileageAt_MissingConfig(t *testing.T) {
	targetTime := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)

	p := &HomeAssistantProvider{apiURL: "", token: "tok", sensorID: "id", client: &http.Client{}}
	_, err := p.GetMileageAt(targetTime)
	if err == nil {
		t.Errorf("Expected error for missing apiURL, got nil")
	}

	p = &HomeAssistantProvider{apiURL: "http://localhost", token: "", sensorID: "id", client: &http.Client{}}
	_, err = p.GetMileageAt(targetTime)
	if err == nil {
		t.Errorf("Expected error for missing token, got nil")
	}

	p = &HomeAssistantProvider{apiURL: "http://localhost", token: "tok", sensorID: "", client: &http.Client{}}
	_, err = p.GetMileageAt(targetTime)
	if err == nil {
		t.Errorf("Expected error for missing sensorID, got nil")
	}
}

func TestHomeAssistantProvider_GetMileageAt_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`this is not json`))
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "test-token", "sensor.test")

	targetTime := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)
	_, err := p.GetMileageAt(targetTime)
	if err == nil {
		t.Fatalf("Expected error for invalid JSON, got nil")
	}
}

// Test that GetMileageAt falls back gracefully when both History and Statistics fail
func TestHomeAssistantProvider_GetMileageAt_FallbackBehavior(t *testing.T) {
	// When History API returns empty and WebSocket connection fails,
	// GetMileageAt should return an error mentioning no data available
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Empty history response - triggers fallback to Statistics
		rw.Write([]byte(`[]`))
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "test-token", "sensor.test")

	targetTime := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)
	_, err := p.GetMileageAt(targetTime)

	// Should fail because:
	// 1. History API returns empty []
	// 2. WebSocket Statistics fallback fails (no real WebSocket server)
	if err == nil {
		t.Fatalf("Expected error when both History and Statistics fail, got nil")
	}

	// Error should mention no data available
	if !strings.Contains(err.Error(), "no data available") {
		t.Errorf("Expected error to mention 'no data available', got: %v", err)
	}
}

// Test getMileageFromHistory directly for more granular testing
func TestHomeAssistantProvider_getMileageFromHistory_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`[[
			{"state": "21350", "attributes": {"unit_of_measurement": "km"}}
		]]`))
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "test-token", "sensor.test_mileage")

	targetTime := time.Date(2026, 2, 28, 23, 59, 59, 0, time.UTC)
	value, err := p.getMileageFromHistory(targetTime)
	if err != nil {
		t.Fatalf("getMileageFromHistory failed: %v", err)
	}

	if value != "21350 km" {
		t.Errorf("Expected '21350 km', got '%s'", value)
	}
}

func TestHomeAssistantProvider_getMileageFromHistory_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`[]`))
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "test-token", "sensor.test")

	targetTime := time.Date(2026, 2, 28, 23, 59, 59, 0, time.UTC)
	_, err := p.getMileageFromHistory(targetTime)
	if err == nil {
		t.Fatalf("Expected error for empty history, got nil")
	}
}

func TestHomeAssistantProvider_getMileageFromHistory_InvalidState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`[[{"state": "unavailable", "attributes": {}}]]`))
	}))
	defer server.Close()

	p := newTestProvider(server.URL, "test-token", "sensor.test")

	targetTime := time.Date(2026, 2, 28, 23, 59, 59, 0, time.UTC)
	_, err := p.getMileageFromHistory(targetTime)
	if err == nil {
		t.Fatalf("Expected error for unavailable state, got nil")
	}
}

// Note: getMileageFromStatistics cannot be easily unit tested because it requires
// a real WebSocket server. Integration tests should be used for this functionality.
// The function is tested indirectly through TestHomeAssistantProvider_GetMileageAt_FallbackBehavior
