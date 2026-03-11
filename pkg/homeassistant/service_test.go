package homeassistant

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewService(t *testing.T) {
	s := NewService("https://homeassistant.local:8123", "mytoken")

	if s == nil {
		t.Error("NewService() returned nil")
	}

	if s.apiURL != "https://homeassistant.local:8123" {
		t.Errorf("expected apiURL 'https://homeassistant.local:8123', got '%s'", s.apiURL)
	}

	if s.token != "mytoken" {
		t.Errorf("expected token 'mytoken', got '%s'", s.token)
	}

	if s.client == nil {
		t.Error("HTTP client should not be nil")
	}
}

func TestNewService_TrimsTrailingSlash(t *testing.T) {
	s := NewService("https://homeassistant.local:8123/", "mytoken")

	if s.apiURL != "https://homeassistant.local:8123" {
		t.Errorf("expected apiURL without trailing slash, got '%s'", s.apiURL)
	}
}

func TestService_GetSensorValue_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify authorization header
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Error("missing Bearer token in Authorization header")
		}

		resp := map[string]interface{}{
			"state": "50000",
			"attributes": map[string]string{
				"unit_of_measurement": "km",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	s := NewService(server.URL, "testtoken")
	value, err := s.GetSensorValue("sensor.car_mileage")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != "50000 km" {
		t.Errorf("expected '50000 km', got '%s'", value)
	}
}

func TestService_GetSensorValue_WithoutUnit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"state":      "on",
			"attributes": map[string]string{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	s := NewService(server.URL, "testtoken")
	value, err := s.GetSensorValue("sensor.switch")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if value != "on" {
		t.Errorf("expected 'on', got '%s'", value)
	}
}

func TestService_GetSensorValue_MissingConfig(t *testing.T) {
	tests := []struct {
		name     string
		apiURL   string
		token    string
		sensorID string
	}{
		{"EmptyAPIURL", "", "token", "sensor.test"},
		{"EmptyToken", "http://localhost", "", "sensor.test"},
		{"EmptySensorID", "http://localhost", "token", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService(tt.apiURL, tt.token)
			value, err := s.GetSensorValue(tt.sensorID)

			if err == nil {
				t.Error("expected error, got nil")
			}

			if value != "unknown" {
				t.Errorf("expected 'unknown', got '%s'", value)
			}

			if !strings.Contains(err.Error(), "missing configuration") {
				t.Errorf("unexpected error message: %v", err)
			}
		})
	}
}

func TestService_GetSensorValue_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Entity not found"))
	}))
	defer server.Close()

	s := NewService(server.URL, "testtoken")
	value, err := s.GetSensorValue("sensor.nonexistent")

	if err == nil {
		t.Error("expected error, got nil")
	}

	if value != "unknown" {
		t.Errorf("expected 'unknown', got '%s'", value)
	}

	if !strings.Contains(err.Error(), "API returned status 404") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestService_GetSensorValue_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	s := NewService(server.URL, "testtoken")
	value, err := s.GetSensorValue("sensor.test")

	if err == nil {
		t.Error("expected error, got nil")
	}

	if value != "unknown" {
		t.Errorf("expected 'unknown', got '%s'", value)
	}

	if !strings.Contains(err.Error(), "failed to parse JSON") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestService_GetSensorValue_UnavailableState(t *testing.T) {
	tests := []struct {
		name  string
		state string
	}{
		{"Unavailable", "unavailable"},
		{"Unknown", "unknown"},
		{"Empty", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resp := map[string]interface{}{
					"state": tt.state,
				}
				json.NewEncoder(w).Encode(resp)
			}))
			defer server.Close()

			s := NewService(server.URL, "testtoken")
			value, err := s.GetSensorValue("sensor.test")

			if err == nil {
				t.Error("expected error, got nil")
			}

			if value != "unknown" {
				t.Errorf("expected 'unknown', got '%s'", value)
			}
		})
	}
}

func TestService_GetSensorValue_RequestFormat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request path
		expectedPath := "/api/states/sensor.test_entity"
		if r.URL.Path != expectedPath {
			t.Errorf("expected path '%s', got '%s'", expectedPath, r.URL.Path)
		}

		// Verify method
		if r.Method != http.MethodGet {
			t.Errorf("expected method GET, got '%s'", r.Method)
		}

		// Verify headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("Content-Type header should be 'application/json'")
		}

		resp := map[string]interface{}{
			"state": "100",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	s := NewService(server.URL, "testtoken")
	_, _ = s.GetSensorValue("sensor.test_entity")
}
