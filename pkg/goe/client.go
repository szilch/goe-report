package goe

import (
	"encoding/json"
	"fmt"
	"goe-report/pkg/config"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/viper"
)

// ChargingLog matches the expected JSON response from the direct_json endpoint
type DirectJsonResp struct {
	Data []ChargingLogRaw `json:"data"`
}

// ChargingLogRaw represents a raw charging log entry as returned by the API
type ChargingLogRaw struct {
	IdChip       interface{} `json:"id_chip"`
	IdChipName   string      `json:"id_chip_name"`
	Start        string      `json:"start"`
	End          string      `json:"end"`
	SecondsTotal string      `json:"seconds_total"`
	Energy       float64     `json:"energy"` // Assumed in kWh
}

// Client handles communication with the go-e API (Cloud or Local).
type Client struct {
	Serial      string
	Token       string
	LocalApiUrl string
	reqUrl      string
}

// NewClient creates a new go-e API client supporting dual environments,
// fetching configuration automatically via viper.
func NewClient() *Client {
	serial := viper.GetString(config.KeySerial)
	token := viper.GetString(config.KeyToken)
	localApiUrl := viper.GetString(config.KeyLocalApiUrl)

	var reqUrl string
	if localApiUrl != "" {
		reqUrl = fmt.Sprintf("%s/api/status", localApiUrl)
	} else {
		reqUrl = fmt.Sprintf("https://%s.api.v3.go-e.io/api/status?token=%s", serial, token)
	}

	return &Client{
		Serial:      serial,
		Token:       token,
		LocalApiUrl: localApiUrl,
		reqUrl:      reqUrl,
	}
}

// GetApiTicket fetches the DLL ticket link and extracts the "e=" parameter.
func (c *Client) GetApiTicket() (string, error) {
	var dllReqUrl string
	if c.LocalApiUrl != "" {
		dllReqUrl = c.reqUrl + "?filter=dll"
	} else {
		dllReqUrl = c.reqUrl + "&filter=dll"
	}

	resp, err := http.Get(dllReqUrl)
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

// GetStatus fetches the current status metrics from the configured go-e API
// and returns a mapped WallboxStatus DTO.
func (c *Client) GetStatus() (*WallboxStatus, error) {
	resp, err := http.Get(c.reqUrl)
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

	var raw rawStatusData
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("error unmarshaling status JSON: %w", err)
	}

	return raw.toDTO(), nil
}
