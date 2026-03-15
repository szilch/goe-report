package carinfo

import (
	"net/http"
	"net/http/httptest"
	"testing"
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

		if req.URL.Path != "/api/states/sensor.test_mileage" {
			t.Errorf("Unexpected path: %s", req.URL.Path)
		}

		rw.Write([]byte(`{
			"state": "50000",
			"attributes": {
				"unit_of_measurement": "km"
			}
		}`))
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
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
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
		rw.Write([]byte(`{"state": "unavailable", "attributes": {}}`))
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
