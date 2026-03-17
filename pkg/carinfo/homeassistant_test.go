package carinfo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestHomeAssistantProvider_GetType(t *testing.T) {
	p := &HomeAssistantProvider{}
	if p.GetType() != "homeassistant" {
		t.Errorf("Expected 'homeassistant', got '%s'", p.GetType())
	}
}

func newTestProviderWS(wsURL, token, sensorID string) *HomeAssistantProvider {
	return &HomeAssistantProvider{
		wsURL:    wsURL,
		token:    token,
		sensorID: sensorID,
		client:   &http.Client{},
	}
}

var upgrader = websocket.Upgrader{}

func setupMockWSServer(t *testing.T, token string, result map[string][]struct {
	Start interface{} `json:"start"`
	End   interface{} `json:"end"`
	Mean  *float64    `json:"mean"`
	State *float64    `json:"state"`
}) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()

		// 1. Send auth_required
		c.WriteJSON(map[string]interface{}{
			"type": "auth_required",
		})

		// 2. Read auth
		var authMsg map[string]interface{}
		if err := c.ReadJSON(&authMsg); err != nil {
			t.Errorf("expected auth message, got %v", err)
			return
		}
		if authMsg["type"] != "auth" || authMsg["access_token"] != token {
			c.WriteJSON(map[string]interface{}{
				"type":    "auth_invalid",
				"message": "Invalid token",
			})
			return
		}

		// 3. Send auth_ok
		c.WriteJSON(map[string]interface{}{
			"type": "auth_ok",
		})

		// 4. Read command
		var cmdMsg map[string]interface{}
		if err := c.ReadJSON(&cmdMsg); err != nil {
			t.Errorf("expected command message, got %v", err)
			return
		}

		if cmdMsg["type"] != "recorder/statistics_during_period" {
			t.Errorf("expected recorder/statistics_during_period, got %v", cmdMsg["type"])
			return
		}

		// 5. Send result
		response := map[string]interface{}{
			"id":      cmdMsg["id"],
			"type":    "result",
			"success": true,
			"result":  result,
		}
		c.WriteJSON(response)
	}))
}

func TestHomeAssistantProvider_GetMileage(t *testing.T) {
	token := "test-token"
	sensorID := "sensor.test_mileage"
	result := map[string][]struct {
		Start interface{} `json:"start"`
		End   interface{} `json:"end"`
		Mean  *float64    `json:"mean"`
		State *float64    `json:"state"`
	}{
		sensorID: {
			{Mean: floatPtrWS(50000.7)},
		},
	}

	server := setupMockWSServer(t, token, result)
	defer server.Close()

	// We pass the full WS URL to the provider as if it came from config
	wsURL := strings.Replace(server.URL, "http://", "ws://", 1) + "/api/websocket"
	p := newTestProviderWS(wsURL, token, sensorID)

	value, err := p.GetMileage()
	if err != nil {
		t.Fatalf("GetMileage failed: %v", err)
	}

	if value != 50000 {
		t.Errorf("Expected 50000 (truncated), got %d", value)
	}
}

func TestHomeAssistantProvider_GetMileageAt(t *testing.T) {
	token := "test-token"
	sensorID := "sensor.test_mileage"
	result := map[string][]struct {
		Start interface{} `json:"start"`
		End   interface{} `json:"end"`
		Mean  *float64    `json:"mean"`
		State *float64    `json:"state"`
	}{
		sensorID: {
			{State: floatPtrWS(48500.0)},
		},
	}

	server := setupMockWSServer(t, token, result)
	defer server.Close()

	wsURL := strings.Replace(server.URL, "http://", "ws://", 1) + "/api/websocket"
	p := newTestProviderWS(wsURL, token, sensorID)

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
	token := "test-token"
	sensorID := "sensor.test"
	result := map[string][]struct {
		Start interface{} `json:"start"`
		End   interface{} `json:"end"`
		Mean  *float64    `json:"mean"`
		State *float64    `json:"state"`
	}{} // Empty result

	server := setupMockWSServer(t, token, result)
	defer server.Close()

	wsURL := strings.Replace(server.URL, "http://", "ws://", 1) + "/api/websocket"
	p := newTestProviderWS(wsURL, token, sensorID)

	targetTime := time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC)
	_, err := p.GetMileageAt(targetTime)
	if err == nil {
		t.Fatalf("Expected error for empty history, got nil")
	}
	if !strings.Contains(err.Error(), "no data for") {
		t.Errorf("Expected error to mention 'no data for', got: %v", err)
	}
}

func floatPtrWS(f float64) *float64 {
	return &f
}

