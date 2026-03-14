package wallbox

import (
	"echarge-report/pkg/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/spf13/viper"
)

// directJsonResp matches the expected JSON response from the direct_json endpoint
type directJsonResp struct {
	Data []chargingLogRaw `json:"data"`
}

// chargingLogRaw represents a raw charging log entry as returned by the go-e API
type chargingLogRaw struct {
	IdChip       interface{} `json:"id_chip"`
	IdChipName   string      `json:"id_chip_name"`
	Start        string      `json:"start"`
	End          string      `json:"end"`
	SecondsTotal string      `json:"seconds_total"`
	Energy       float64     `json:"energy"` // in kWh
}

// rawStatusData contains the raw JSON structure of the go-e Charger status API response.
type rawStatusData struct {
	Car int       `json:"car"` // 1: idle, 2: charging, 3: wait car, 4: complete, 5: error
	Alw bool      `json:"alw"`
	Amp int       `json:"amp"`
	Wh  float64   `json:"wh"`
	Eto float64   `json:"eto"`
	Nrg []float64 `json:"nrg"`
	Tma []float64 `json:"tma"`
	Frc int       `json:"frc"`
}

// goeAdapter implements the Adapter interface for go-e chargers.
type goeAdapter struct {
	Serial        string
	Token         string
	LocalApiUrl   string
	reqUrl        string
	directJsonUrl string
}

// newGoeAdapter creates a new go-e wallbox adapter.
// Configuration is fetched automatically via viper.
func newGoeAdapter() *goeAdapter {
	serial := viper.GetString(config.KeyWallboxGoeCloudSerial)
	token := viper.GetString(config.KeyWallboxGoeCloudToken)
	localApiUrl := viper.GetString(config.KeyWallboxGoeLocalApiUrl)

	var reqUrl string
	if localApiUrl != "" {
		reqUrl = fmt.Sprintf("%s/api/status", localApiUrl)
	} else {
		reqUrl = fmt.Sprintf("https://%s.api.v3.go-e.io/api/status?token=%s", serial, token)
	}

	return &goeAdapter{
		Serial:        serial,
		Token:         token,
		LocalApiUrl:   localApiUrl,
		reqUrl:        reqUrl,
		directJsonUrl: "https://data.v3.go-e.io/api/v1/direct_json",
	}
}

// GetType returns the type identifier of this adapter.
func (a *goeAdapter) GetType() string {
	return "goe"
}

// getApiTicket fetches the DLL ticket link and extracts the "e=" parameter.
func (a *goeAdapter) getApiTicket() (string, error) {
	parsed, err := url.Parse(a.reqUrl)
	if err != nil {
		return "", fmt.Errorf("invalid reqUrl: %w", err)
	}
	q := parsed.Query()
	q.Set("filter", "dll")
	parsed.RawQuery = q.Encode()
	dllReqUrl := parsed.String()

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

// FetchChargingData fetches the charging data for a given timeframe and converts it to the generic format.
func (a *goeAdapter) FetchChargingData(fromMs, toMs int64) (*ChargingResponse, error) {
	ticket, err := a.getApiTicket()
	if err != nil {
		return nil, fmt.Errorf("error getting API ticket: %w", err)
	}

	jsonUrl := fmt.Sprintf("%s?e=%s&from=%d&to=%d&timezone=Europe/Berlin", a.directJsonUrl, ticket, fromMs, toMs)

	jsonResp, err := http.Get(jsonUrl)
	if err != nil {
		return nil, fmt.Errorf("error fetching JSON charging data: %w", err)
	}
	defer jsonResp.Body.Close()

	jsonBody, err := io.ReadAll(jsonResp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading JSON data: %w", err)
	}

	var responseData directJsonResp
	if err := json.Unmarshal(jsonBody, &responseData); err != nil {
		return nil, fmt.Errorf("error parsing charging data: %w", err)
	}

	// Convert to generic format
	return a.toChargingResponse(&responseData), nil
}

// toChargingResponse converts the go-e specific response to the generic ChargingResponse.
func (a *goeAdapter) toChargingResponse(data *directJsonResp) *ChargingResponse {
	sessions := make([]ChargingSession, len(data.Data))
	for i, raw := range data.Data {
		sessions[i] = ChargingSession{
			IdChip:       raw.IdChip,
			IdChipName:   raw.IdChipName,
			Start:        raw.Start,
			End:          raw.End,
			SecondsTotal: raw.SecondsTotal,
			Energy:       raw.Energy,
		}
	}
	return &ChargingResponse{Data: sessions}
}

// GetStatus fetches the current status metrics from the configured go-e API
// and returns a generalized Status DTO.
func (a *goeAdapter) GetStatus() (*Status, error) {
	resp, err := http.Get(a.reqUrl)
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

	return a.toStatus(&raw), nil
}

// toStatus converts the go-e specific raw status data to the generic Status.
func (a *goeAdapter) toStatus(raw *rawStatusData) *Status {
	status := &Status{
		SetCurrentA:            raw.Amp,
		ChargedSincePlugInKWh:  raw.Wh / 1000.0,
		TotalEnergyLifetimeKWh: raw.Eto / 1000.0,
	}

	// Interpret the 'car' state
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

	// Allowed state
	status.ChargingAllowed = "No"
	if raw.Alw {
		status.ChargingAllowed = "Yes"
	}

	// Calculate total power
	numNrg := len(raw.Nrg)
	if numNrg >= 12 {
		status.CurrentPowerKW = raw.Nrg[11] / 1000.0
	}

	// Temperature
	status.TemperatureCelsius = "N/A"
	if len(raw.Tma) > 0 {
		status.TemperatureCelsius = fmt.Sprintf("%.1f °C", raw.Tma[0])
	}

	// Phase details
	if numNrg >= 10 {
		status.Phases = []PhaseDetail{
			{Voltage: raw.Nrg[0], Current: raw.Nrg[4], Power: raw.Nrg[7]},
			{Voltage: raw.Nrg[1], Current: raw.Nrg[5], Power: raw.Nrg[8]},
			{Voltage: raw.Nrg[2], Current: raw.Nrg[6], Power: raw.Nrg[9]},
		}
	}

	return status
}
