package goe

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client handles communication with the go-e API (Cloud or Local).
type Client struct {
	Serial      string
	Token       string
	LocalApiUrl string
}

// NewClient creates a new go-e API client supporting dual environments.
func NewClient(serial, token, localApiUrl string) *Client {
	return &Client{
		Serial:      serial,
		Token:       token,
		LocalApiUrl: localApiUrl,
	}
}

// GetApiTicket fetches the DLL ticket link and extracts the "e=" parameter.
func (c *Client) GetApiTicket() (string, error) {
	var reqUrl string
	if c.LocalApiUrl != "" {
		reqUrl = fmt.Sprintf("%s/api/status?filter=dll", c.LocalApiUrl)
	} else {
		reqUrl = fmt.Sprintf("https://%s.api.v3.go-e.io/api/status?token=%s&filter=dll", c.Serial, c.Token)
	}

	resp, err := http.Get(reqUrl)
	if err != nil {
		return "", fmt.Errorf("error fetching API ticket: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response %w", err)
	}

	var dllResp struct {
		Dll string `json:"dll"`
	}
	if err := json.Unmarshal(body, &dllResp); err != nil {
		return "", fmt.Errorf("error parsing API response: %w", err)
	}

	if dllResp.Dll == "" {
		return "", fmt.Errorf("could not obtain a ticket from the API")
	}

	parsedUrl, err := url.Parse(dllResp.Dll)
	if err != nil {
		return "", fmt.Errorf("error parsing URL: %w", err)
	}

	ticket := parsedUrl.Query().Get("e")
	if ticket == "" {
		return "", fmt.Errorf("could not extract ticket from URL")
	}

	return ticket, nil
}

// FetchChargingData fetches the direct JSON charging data for a given timeframe using the provided ticket.
func (c *Client) FetchChargingData(ticket string, fromMs, toMs int64) (*DirectJsonResp, error) {
	jsonUrl := fmt.Sprintf("https://data.v3.go-e.io/api/v1/direct_json?e=%s&from=%d&to=%d&timezone=Europe/Berlin", ticket, fromMs, toMs)

	jsonResp, err := http.Get(jsonUrl)
	if err != nil {
		return nil, fmt.Errorf("error fetching JSON charging data: %w", err)
	}
	defer jsonResp.Body.Close()

	jsonBody, err := io.ReadAll(jsonResp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading JSON data: %w", err)
	}

	var responseData DirectJsonResp
	if err := json.Unmarshal(jsonBody, &responseData); err != nil {
		return nil, fmt.Errorf("error parsing charging data: %w", err)
	}

	return &responseData, nil
}

// StatusData contains the structured response of the go-e Charger status API.
type StatusData struct {
	Car int       `json:"car"` // 1: idle, 2: charging, 3: wait car, 4: complete, 5: error
	Alw bool      `json:"alw"`
	Amp int       `json:"amp"`
	Wh  float64   `json:"wh"`
	Eto float64   `json:"eto"`
	Nrg []float64 `json:"nrg"`
	Tma []float64 `json:"tma"`
	Frc int       `json:"frc"`
}

// GetStatus fetchs the current status metrics from the configured go-e API.
func (c *Client) GetStatus() (*StatusData, error) {
	var reqUrl string
	if c.LocalApiUrl != "" {
		reqUrl = fmt.Sprintf("%s/api/status", c.LocalApiUrl)
	} else {
		reqUrl = fmt.Sprintf("https://%s.api.v3.go-e.io/api/status?token=%s", c.Serial, c.Token)
	}

	resp, err := http.Get(reqUrl)
	if err != nil {
		return nil, fmt.Errorf("error connecting to the go-e API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API responded with status %d: %s", resp.StatusCode, string(body))
	}

	var statusData StatusData
	if err := json.Unmarshal(body, &statusData); err != nil {
		return nil, fmt.Errorf("error unmarshaling status JSON: %w", err)
	}

	return &statusData, nil
}
