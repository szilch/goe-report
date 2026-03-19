package carinfo

import (
	"fmt"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// HomeAssistantProvider implements the Provider interface for Home Assistant.
type HomeAssistantProvider struct {
	wsURL    string
	token    string
	sensorID string
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

// NewHomeAssistantProvider creates a new HomeAssistantProvider using the provided config.
func NewHomeAssistantProvider(cfg Config) *HomeAssistantProvider {
	wsURL, _ := url.JoinPath(cfg.HAWsHost, "/api/websocket")

	return &HomeAssistantProvider{
		wsURL:    wsURL,
		token:    cfg.HAToken,
		sensorID: cfg.HAMileageSensor,
	}
}

// GetType returns the provider type identifier for Home Assistant.
func (p *HomeAssistantProvider) GetType() string {
	return TypeHomeAssistant
}

// GetMileage returns the current mileage by fetching the state at the current time.
func (p *HomeAssistantProvider) GetMileage() (int, error) {
	return p.GetMileageAt(time.Now())
}

// GetMileageAt fetches the mileage for a specific date using the Home Assistant WebSocket API.
func (p *HomeAssistantProvider) GetMileageAt(t time.Time) (int, error) {
	c, _, err := websocket.DefaultDialer.Dial(p.wsURL, nil)
	if err != nil {
		return 0, fmt.Errorf("dial websocket: %w", err)
	}
	defer c.Close()

	var msg map[string]interface{}
	if err := c.ReadJSON(&msg); err != nil || msg["type"] != "auth_required" {
		return 0, fmt.Errorf("auth_required message not received")
	}

	c.WriteJSON(map[string]string{
		"type":         "auth",
		"access_token": p.token,
	})

	if err := c.ReadJSON(&msg); err != nil || msg["type"] != "auth_ok" {
		return 0, fmt.Errorf("authenticate: %v", msg["message"])
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
		return 0, fmt.Errorf("read response: %w", err)
	}

	if !resp.Success {
		return 0, fmt.Errorf("API error: %s - %s", resp.Error.Code, resp.Error.Message)
	}

	dataList, ok := resp.Result[p.sensorID]
	if !ok || len(dataList) == 0 {
		return 0, fmt.Errorf("%w: %s at %s", ErrNoData, p.sensorID, t)
	}

	if dataList[0].Mean != nil {
		return int(*dataList[0].Mean), nil
	} else if dataList[0].State != nil {
		return int(*dataList[0].State), nil
	}

	return 0, fmt.Errorf("no data found")
}
