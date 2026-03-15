package homeassistant

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestService(serverURL, token string) *Service {
	return &Service{
		apiURL: serverURL,
		token:  token,
		client: &http.Client{},
	}
}

func TestService_GetSensorValue(t *testing.T) {
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

	s := newTestService(server.URL, "test-token")

	value, err := s.GetSensorValue("sensor.test_mileage")
	if err != nil {
		t.Fatalf("GetSensorValue failed: %v", err)
	}

	if value != "50000 km" {
		t.Errorf("Expected '50000 km', got '%s'", value)
	}
}

func TestService_GetSensorValue_NoUnit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`{"state": "42", "attributes": {}}`))
	}))
	defer server.Close()

	s := newTestService(server.URL, "test-token")

	value, err := s.GetSensorValue("sensor.test")
	if err != nil {
		t.Fatalf("GetSensorValue failed: %v", err)
	}

	if value != "42" {
		t.Errorf("Expected '42', got '%s'", value)
	}
}

func TestService_GetSensorValue_UnavailableState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`{"state": "unavailable", "attributes": {}}`))
	}))
	defer server.Close()

	s := newTestService(server.URL, "test-token")

	_, err := s.GetSensorValue("sensor.test")
	if err == nil {
		t.Fatalf("Expected error for unavailable state, got nil")
	}
}

func TestService_GetSensorValue_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusUnauthorized)
		rw.Write([]byte(`{"message": "401: Unauthorized"}`))
	}))
	defer server.Close()

	s := newTestService(server.URL, "wrong-token")

	_, err := s.GetSensorValue("sensor.test")
	if err == nil {
		t.Fatalf("Expected error for HTTP 401, got nil")
	}
}

func TestService_GetSensorValue_MissingConfig(t *testing.T) {
	s := &Service{apiURL: "", token: "tok", client: &http.Client{}}
	_, err := s.GetSensorValue("sensor.test")
	if err == nil {
		t.Errorf("Expected error for missing apiURL, got nil")
	}

	s = &Service{apiURL: "http://localhost", token: "", client: &http.Client{}}
	_, err = s.GetSensorValue("sensor.test")
	if err == nil {
		t.Errorf("Expected error for missing token, got nil")
	}

	s = &Service{apiURL: "http://localhost", token: "tok", client: &http.Client{}}
	_, err = s.GetSensorValue("")
	if err == nil {
		t.Errorf("Expected error for missing sensorID, got nil")
	}
}

func TestService_GetSensorValue_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`this is not json`))
	}))
	defer server.Close()

	s := newTestService(server.URL, "test-token")

	_, err := s.GetSensorValue("sensor.test")
	if err == nil {
		t.Fatalf("Expected error for invalid JSON, got nil")
	}
}
