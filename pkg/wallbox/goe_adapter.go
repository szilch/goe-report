package wallbox

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"echarge-report/pkg/config"

	"github.com/spf13/viper"
)

type directJsonResp struct {
	Data []chargingLogRaw `json:"data"`
}

type chargingLogRaw struct {
	IdChip       interface{} `json:"id_chip"`
	IdChipName   string      `json:"id_chip_name"`
	Start        string      `json:"start"`
	End          string      `json:"end"`
	SecondsTotal string      `json:"seconds_total"`
	Energy       float64     `json:"energy"`
}

type rawStatusData struct {
	Car int       `json:"car"`
	Alw bool      `json:"alw"`
	Amp int       `json:"amp"`
	Wh  float64   `json:"wh"`
	Eto float64   `json:"eto"`
	Nrg []float64 `json:"nrg"`
	Tma []float64 `json:"tma"`
	Frc int       `json:"frc"`
}

type goeAdapter struct {
	serial        string
	token         string
	localAPIURL   string
	reqURL        string
	directJSONURL string
}

func newGoeAdapter() *goeAdapter {
	serial := viper.GetString(config.KeyWallboxGoeCloudSerial)
	token := viper.GetString(config.KeyWallboxGoeCloudToken)
	localAPIURL := viper.GetString(config.KeyWallboxGoeLocalApiUrl)

	var reqURL string
	if localAPIURL != "" {
		reqURL = fmt.Sprintf("%s/api/status", localAPIURL)
	} else {
		reqURL = fmt.Sprintf("https://%s.api.v3.go-e.io/api/status?token=%s", serial, token)
	}

	return &goeAdapter{
		serial:        serial,
		token:         token,
		localAPIURL:   localAPIURL,
		reqURL:        reqURL,
		directJSONURL: "https://data.v3.go-e.io/api/v1/direct_json",
	}
}

func (a *goeAdapter) GetType() string {
	return TypeGoE
}

func (a *goeAdapter) getAPITicket() (string, error) {
	parsed, err := url.Parse(a.reqURL)
	if err != nil {
		return "", fmt.Errorf("parse request URL: %w", err)
	}
	q := parsed.Query()
	q.Set("filter", "dll")
	parsed.RawQuery = q.Encode()
	dllReqURL := parsed.String()

	resp, err := http.Get(dllReqURL)
	if err != nil {
		return "", fmt.Errorf("fetch API ticket: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	var dllResp struct {
		Dll string `json:"dll"`
	}
	if err := json.Unmarshal(body, &dllResp); err != nil {
		return "", fmt.Errorf("parse API response: %w", err)
	}

	if dllResp.Dll == "" {
		return "", fmt.Errorf("could not obtain a ticket from the API")
	}

	parsedURL, err := url.Parse(dllResp.Dll)
	if err != nil {
		return "", fmt.Errorf("parse ticket URL: %w", err)
	}

	ticket := parsedURL.Query().Get("e")
	if ticket == "" {
		return "", fmt.Errorf("could not extract ticket from URL")
	}

	return ticket, nil
}

func (a *goeAdapter) FetchChargingData(fromMs, toMs int64) (*ChargingResponse, error) {
	ticket, err := a.getAPITicket()
	if err != nil {
		return nil, fmt.Errorf("get API ticket: %w", err)
	}

	jsonURL := fmt.Sprintf("%s?e=%s&from=%d&to=%d&timezone=Europe/Berlin", a.directJSONURL, ticket, fromMs, toMs)

	jsonResp, err := http.Get(jsonURL)
	if err != nil {
		return nil, fmt.Errorf("fetch JSON charging data: %w", err)
	}
	defer jsonResp.Body.Close()

	jsonBody, err := io.ReadAll(jsonResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read JSON data: %w", err)
	}

	var responseData directJsonResp
	if err := json.Unmarshal(jsonBody, &responseData); err != nil {
		return nil, fmt.Errorf("parse charging data: %w", err)
	}

	return a.toChargingResponse(&responseData), nil
}

func (a *goeAdapter) toChargingResponse(data *directJsonResp) *ChargingResponse {
	sessions := make([]ChargingSession, len(data.Data))
	for i, raw := range data.Data {
		sessions[i] = ChargingSession{
			IdChip:     raw.IdChip,
			IdChipName: raw.IdChipName,
			Start:      a.parseTime(raw.Start),
			End:        a.parseTime(raw.End),
			Duration:   a.parseDuration(raw.SecondsTotal),
			Energy:     raw.Energy,
		}
	}
	return &ChargingResponse{Data: sessions}
}

func (a *goeAdapter) parseDuration(durationStr string) time.Duration {
	var h, m, s int
	if _, err := fmt.Sscanf(durationStr, "%d:%d:%d", &h, &m, &s); err != nil {
		return 0
	}
	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(s)*time.Second
}

func (a *goeAdapter) parseTime(timeStr string) time.Time {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		// Fall back to UTC if the timezone database is unavailable.
		loc = time.UTC
	}

	layout := "02.01.2006 15:04:05"
	if t, err := time.ParseInLocation(layout, timeStr, time.UTC); err == nil {
		return t.In(loc)
	}

	return time.Time{}
}

func (a *goeAdapter) GetStatus() (*Status, error) {
	resp, err := http.Get(a.reqURL)
	if err != nil {
		return nil, fmt.Errorf("connect to go-e API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API responded with status %d: %s", resp.StatusCode, string(body))
	}

	var raw rawStatusData
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal status JSON: %w", err)
	}

	return a.toStatus(&raw), nil
}

func (a *goeAdapter) toStatus(raw *rawStatusData) *Status {
	status := &Status{
		SetCurrentA:            raw.Amp,
		ChargedSincePlugInKWh:  raw.Wh / 1000.0,
		TotalEnergyLifetimeKWh: raw.Eto / 1000.0,
	}

	switch raw.Car {
	case 1:
		status.VehicleState = "Idle (not connected)"
	case 2:
		status.VehicleState = "Charging"
	case 3:
		status.VehicleState = "Waiting for car"
	case 4:
		status.VehicleState = "Charging complete"
	case 5:
		status.VehicleState = "Error"
	default:
		status.VehicleState = "Unknown"
	}

	status.ChargingAllowed = raw.Alw

	numNrg := len(raw.Nrg)
	if numNrg >= 12 {
		status.CurrentPowerKW = raw.Nrg[11] / 1000.0
	}

	status.TemperatureCelsius = "N/A"
	if len(raw.Tma) > 0 {
		status.TemperatureCelsius = fmt.Sprintf("%.1f °C", raw.Tma[0])
	}

	if numNrg >= 10 {
		status.Phases = []PhaseDetail{
			{Voltage: raw.Nrg[0], Current: raw.Nrg[4], Power: raw.Nrg[7]},
			{Voltage: raw.Nrg[1], Current: raw.Nrg[5], Power: raw.Nrg[8]},
			{Voltage: raw.Nrg[2], Current: raw.Nrg[6], Power: raw.Nrg[9]},
		}
	}

	return status
}
