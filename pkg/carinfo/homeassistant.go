package carinfo

import (
	"crypto/tls"
	"echarge-report/pkg/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
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
	LastChanged string `json:"last_changed"`
}

func (p *HomeAssistantProvider) GetType() string {
	return TypeHomeAssistant
}

func (p *HomeAssistantProvider) GetMileage() (string, error) {
	return p.GetMileageAt(time.Now())
}

// GetMileageAt returns the mileage at a specific point in time.
// It first tries the History API (for recent data within ~10 days),
// then falls back to Long-Term Statistics via WebSocket (for older data).
func (p *HomeAssistantProvider) GetMileageAt(t time.Time) (string, error) {
	if p.apiURL == "" || p.token == "" || p.sensorID == "" {
		return "unknown", fmt.Errorf("missing configuration: apiURL=%q, token_set=%v, sensorID=%q", p.apiURL, p.token != "", p.sensorID)
	}

	// First try History API (works for recent data)
	result, err := p.getMileageFromHistory(t)
	if err == nil && result != "" && result != "unknown" {
		return result, nil
	}

	// Fall back to Long-Term Statistics via WebSocket
	result, err = p.getMileageFromStatistics(t)
	if err == nil && result != "" && result != "unknown" {
		return result, nil
	}

	// If statistics also failed, return the history error or statistics error
	if err != nil {
		return "unknown", fmt.Errorf("no data available for %s: %w", t.Format("02.01.2006"), err)
	}
	return "unknown", fmt.Errorf("no data available for %s", t.Format("02.01.2006"))
}

// getMileageFromHistory tries to get mileage from the History API (recent data only)
func (p *HomeAssistantProvider) getMileageFromHistory(t time.Time) (string, error) {
	startTime := t.Add(-7 * 24 * time.Hour).UTC().Format(time.RFC3339)
	endTime := t.UTC().Format(time.RFC3339)

	reqURL := fmt.Sprintf("%s/api/history/period/%s?filter_entity_id=%s&end_time=%s",
		p.apiURL,
		url.PathEscape(startTime),
		url.QueryEscape(p.sensorID),
		url.QueryEscape(endTime))

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create history request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("history connection error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("history API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read history response: %w", err)
	}

	var historyResponse [][]stateResponse
	if err := json.Unmarshal(body, &historyResponse); err != nil {
		return "", fmt.Errorf("failed to parse history JSON: %w", err)
	}

	if len(historyResponse) == 0 || len(historyResponse[0]) == 0 {
		return "", fmt.Errorf("no history data")
	}

	states := historyResponse[0]
	state := states[len(states)-1]

	if state.State == "" || state.State == "unavailable" || state.State == "unknown" {
		return "", fmt.Errorf("invalid state: %q", state.State)
	}

	unit := state.Attributes.UnitOfMeasurement
	if unit == "" {
		unit, _ = p.getUnitOfMeasurement()
	}

	if unit != "" {
		return fmt.Sprintf("%s %s", state.State, unit), nil
	}
	return state.State, nil
}

// getMileageFromStatistics gets mileage from Long-Term Statistics via WebSocket API
func (p *HomeAssistantProvider) getMileageFromStatistics(t time.Time) (string, error) {
	// Convert HTTP URL to WebSocket URL
	wsURL := strings.Replace(p.apiURL, "https://", "wss://", 1)
	wsURL = strings.Replace(wsURL, "http://", "ws://", 1)
	wsURL = wsURL + "/api/websocket"

	// Setup WebSocket connection with TLS config
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		return "", fmt.Errorf("websocket connection failed: %w", err)
	}
	defer conn.Close()

	// Read auth_required message
	var authRequired map[string]interface{}
	if err := conn.ReadJSON(&authRequired); err != nil {
		return "", fmt.Errorf("failed to read auth_required: %w", err)
	}

	// Send auth message
	authMsg := map[string]string{
		"type":         "auth",
		"access_token": p.token,
	}
	if err := conn.WriteJSON(authMsg); err != nil {
		return "", fmt.Errorf("failed to send auth: %w", err)
	}

	// Read auth result
	var authResult map[string]interface{}
	if err := conn.ReadJSON(&authResult); err != nil {
		return "", fmt.Errorf("failed to read auth result: %w", err)
	}
	if authResult["type"] != "auth_ok" {
		return "", fmt.Errorf("authentication failed: %v", authResult)
	}

	// Query statistics for the end of the target day
	// We request statistics from the start of the day to get the value at that point
	startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	statsMsg := map[string]interface{}{
		"id":            1,
		"type":          "recorder/statistics_during_period",
		"start_time":    startOfDay.UTC().Format(time.RFC3339),
		"end_time":      endOfDay.UTC().Format(time.RFC3339),
		"statistic_ids": []string{p.sensorID},
		"period":        "hour",
	}
	if err := conn.WriteJSON(statsMsg); err != nil {
		return "", fmt.Errorf("failed to send statistics request: %w", err)
	}

	// Read statistics response
	var statsResult map[string]interface{}
	if err := conn.ReadJSON(&statsResult); err != nil {
		return "", fmt.Errorf("failed to read statistics response: %w", err)
	}

	if !statsResult["success"].(bool) {
		return "", fmt.Errorf("statistics request failed: %v", statsResult)
	}

	// Parse the result
	result, ok := statsResult["result"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid statistics result format")
	}

	sensorStats, ok := result[p.sensorID].([]interface{})
	if !ok || len(sensorStats) == 0 {
		return "", fmt.Errorf("no statistics data for sensor")
	}

	// Get the last statistics entry (closest to end of day)
	lastStat := sensorStats[len(sensorStats)-1].(map[string]interface{})

	// Try to get the "state" or "max" value
	var value float64
	if state, ok := lastStat["state"].(float64); ok {
		value = state
	} else if max, ok := lastStat["max"].(float64); ok {
		value = max
	} else if mean, ok := lastStat["mean"].(float64); ok {
		value = mean
	} else {
		return "", fmt.Errorf("no state/max/mean in statistics")
	}

	// Get unit from current state
	unit, _ := p.getUnitOfMeasurement()
	if unit != "" {
		return fmt.Sprintf("%.0f %s", value, unit), nil
	}
	return fmt.Sprintf("%.0f", value), nil
}

// getUnitOfMeasurement fetches the unit of measurement from the current state
func (p *HomeAssistantProvider) getUnitOfMeasurement() (string, error) {
	reqURL := fmt.Sprintf("%s/api/states/%s", p.apiURL, p.sensorID)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var state stateResponse
	if err := json.Unmarshal(body, &state); err != nil {
		return "", err
	}

	return state.Attributes.UnitOfMeasurement, nil
}
