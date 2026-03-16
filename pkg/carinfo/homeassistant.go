package carinfo

import (
	"crypto/tls"
	"echarge-report/pkg/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

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
	State string `json:"state"`
}

func (p *HomeAssistantProvider) GetType() string {
	return TypeHomeAssistant
}

func (p *HomeAssistantProvider) GetMileage() (int, error) {
	return p.GetMileageAt(time.Now())
}

func (p *HomeAssistantProvider) GetMileageAt(t time.Time) (int, error) {
	if p.apiURL == "" || p.token == "" || p.sensorID == "" {
		return 0, fmt.Errorf("missing configuration: apiURL=%q, token_set=%v, sensorID=%q", p.apiURL, p.token != "", p.sensorID)
	}
	tsStr := t.Format(time.RFC3339)
	fullURL := fmt.Sprintf("%s/api/history/period/%s?filter_entity_id=%s&end_time=%s&minimal_response",
		p.apiURL,
		url.QueryEscape(tsStr),
		p.sensorID,
		url.QueryEscape(tsStr),
	)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", fullURL, nil)
	req.Header.Set("Authorization", "Bearer "+p.token)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Home Assistant antwortet mit [][]State
	var history [][]stateResponse
	if err := json.Unmarshal(body, &history); err != nil {
		return 0, err
	}

	// Prüfen, ob wir Daten erhalten haben
	if len(history) > 0 && len(history[0]) > 0 {
		return strconv.Atoi(history[0][0].State)
	}

	return 0, fmt.Errorf("keine Daten für diesen Zeitpunkt gefunden")
}
