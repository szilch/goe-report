package carinfo

import (
	"crypto/tls"
	"echarge-report/pkg/config"
	"fmt"
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

type wsResponse struct {
	ID      int    `json:"id"`
	Type    string `json:"type"`
	Success bool   `json:"success"`
	Result  map[string][]struct {
		Start interface{} `json:"start"`
		End   interface{} `json:"end"`
		Mean  *float64    `json:"mean"`
		State *float64    `json:"state"`
	} `json:"result"`
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func NewHomeAssistantProvider() *HomeAssistantProvider {
	apiURL := viper.GetString(config.KeyHAWsHost)
	token := viper.GetString(config.KeyHAToken)
	sensorID := viper.GetString(config.KeyHAMilageSensor)

	tlsCfg := &tls.Config{}
	transport := &http.Transport{TLSClientConfig: tlsCfg}
	return &HomeAssistantProvider{
		apiURL:   strings.TrimRight(apiURL, "/"),
		token:    token,
		sensorID: sensorID,
		client:   &http.Client{Transport: transport},
	}
}

func (p *HomeAssistantProvider) GetType() string {
	return TypeHomeAssistant
}

func (p *HomeAssistantProvider) GetMileage() (int, error) {
	return p.GetMileageAt(time.Now())
}

func (p *HomeAssistantProvider) GetMileageAt(t time.Time) (int, error) {
	wsUrl, err := convertToWsUrl(p.apiURL)
	if err != nil {
		return 0, fmt.Errorf("error generating ws url: %v", err)
	}
	c, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		return 0, fmt.Errorf("dial error: %v", err)
	}
	defer c.Close()

	var msg map[string]interface{}
	if err := c.ReadJSON(&msg); err != nil || msg["type"] != "auth_required" {
		return 0, fmt.Errorf("auth_required not received")
	}

	c.WriteJSON(map[string]string{
		"type":         "auth",
		"access_token": p.token,
	})

	if err := c.ReadJSON(&msg); err != nil || msg["type"] != "auth_ok" {
		return 0, fmt.Errorf("authentication failed: %v", msg["message"])
	}
	startOfDay := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)

	c.WriteJSON(map[string]interface{}{
		"id":            1,
		"type":          "recorder/statistics_during_period",
		"start_time":    startOfDay.Format(time.RFC3339),
		"end_time":      endOfDay.Format(time.RFC3339),
		"period":        "day",
		"statistic_ids": []string{p.sensorID},
		"types":         []string{"mean", "state"},
	})

	var resp wsResponse
	if err := c.ReadJSON(&resp); err != nil {
		return 0, fmt.Errorf("error reading response: %v", err)
	}

	if !resp.Success {
		return 0, fmt.Errorf("API error: %s - %s", resp.Error.Code, resp.Error.Message)
	}

	data, ok := resp.Result[p.sensorID]
	if !ok || len(data) == 0 {
		return 0, fmt.Errorf("no data for %s at %s", p.sensorID, t)
	}

	if data[0].Mean != nil {
		return int(*data[0].Mean), nil
	} else if data[0].State != nil {
		return int(*data[0].State), nil
	}

	return 0, fmt.Errorf("no data found")
}

func convertToWsUrl(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}
	u.Path, _ = url.JoinPath(u.Path, "/api/websocket")

	return u.String(), nil
}
