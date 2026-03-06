package homeassistant

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Service is a generic client for the Home Assistant REST API.
// SSL certificate verification is intentionally disabled.
type Service struct {
	apiURL string
	token  string
	client *http.Client
}

// NewService creates a new HA service with the given API URL and token.
func NewService(apiURL, token string) *Service {
	tlsCfg := &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	transport := &http.Transport{TLSClientConfig: tlsCfg}
	return &Service{
		apiURL: strings.TrimRight(apiURL, "/"),
		token:  token,
		client: &http.Client{Transport: transport},
	}
}

// stateResponse maps the relevant fields of the HA /api/states/<entity_id> response.
type stateResponse struct {
	State      string `json:"state"`
	Attributes struct {
		UnitOfMeasurement string `json:"unit_of_measurement"`
	} `json:"attributes"`
}

// GetSensorValue fetches the current state of the given sensor.
// Returns the value (with unit if available) as a string, and an error if it fails.
func (s *Service) GetSensorValue(sensorID string) (string, error) {
	if s.apiURL == "" || s.token == "" || sensorID == "" {
		return "unknown", fmt.Errorf("missing configuration: apiURL=%q, token_set=%v, sensorID=%q", s.apiURL, s.token != "", sensorID)
	}

	reqURL := fmt.Sprintf("%s/api/states/%s", s.apiURL, sensorID)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return "unknown", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "unknown", fmt.Errorf("connection error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "unknown", fmt.Errorf("API returned status %d. Body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "unknown", fmt.Errorf("failed to read response body: %w", err)
	}

	var state stateResponse
	if err := json.Unmarshal(body, &state); err != nil {
		return "unknown", fmt.Errorf("failed to parse JSON: %w (body: %s)", err, string(body))
	}

	if state.State == "" || state.State == "unavailable" || state.State == "unknown" {
		return "unknown", fmt.Errorf("sensor reported unavailable/unknown state: %q", state.State)
	}

	if state.Attributes.UnitOfMeasurement != "" {
		return fmt.Sprintf("%s %s", state.State, state.Attributes.UnitOfMeasurement), nil
	}
	return state.State, nil
}
