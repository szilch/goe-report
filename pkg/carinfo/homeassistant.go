package carinfo

import (
	"crypto/tls"
	"echarge-report/pkg/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/viper"
)

type HomeAssistantProvider struct {
	apiURL   string
	token    string
	sensorID string
	client   *http.Client
}

func NewHomeAssistantProvider() *HomeAssistantProvider {
	apiURL := viper.GetString(config.KeyHAAPI)
	token := viper.GetString(config.KeyHAToken)
	sensorID := viper.GetString(config.KeyHAMilageSensor)

	tlsCfg := &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	transport := &http.Transport{TLSClientConfig: tlsCfg}
	return &HomeAssistantProvider{
		apiURL:   strings.TrimRight(apiURL, "/"),
		token:    token,
		sensorID: sensorID,
		client:   &http.Client{Transport: transport},
	}
}

type stateResponse struct {
	State      string `json:"state"`
	Attributes struct {
		UnitOfMeasurement string `json:"unit_of_measurement"`
	} `json:"attributes"`
}

func (p *HomeAssistantProvider) GetType() string {
	return TypeHomeAssistant
}

func (p *HomeAssistantProvider) GetMileage() (string, error) {
	if p.apiURL == "" || p.token == "" || p.sensorID == "" {
		return "unknown", fmt.Errorf("missing configuration: apiURL=%q, token_set=%v, sensorID=%q", p.apiURL, p.token != "", p.sensorID)
	}

	reqURL := fmt.Sprintf("%s/api/states/%s", p.apiURL, p.sensorID)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return "unknown", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
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
