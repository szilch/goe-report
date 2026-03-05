package ha

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Service ist ein generischer Client für die Home Assistant REST API.
// SSL-Zertifikatsprüfung ist bewusst deaktiviert.
type Service struct {
	apiURL string
	token  string
	client *http.Client
}

// NewService erstellt einen neuen HA-Service mit dem gegebenen API-URL und Token.
func NewService(apiURL, token string) *Service {
	tlsCfg := &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	transport := &http.Transport{TLSClientConfig: tlsCfg}
	return &Service{
		apiURL: strings.TrimRight(apiURL, "/"),
		token:  token,
		client: &http.Client{Transport: transport},
	}
}

// stateResponse bildet die relevanten Felder der HA /api/states/<entity_id> Antwort ab.
type stateResponse struct {
	State      string `json:"state"`
	Attributes struct {
		UnitOfMeasurement string `json:"unit_of_measurement"`
	} `json:"attributes"`
}

// GetSensorValue ruft den aktuellen Zustand des angegebenen Sensors ab.
// Gibt den Wert (ggf. mit Einheit) als String zurück, oder "unbekannt" bei Fehler.
func (s *Service) GetSensorValue(sensorID string) string {
	if s.apiURL == "" || s.token == "" || sensorID == "" {
		return "unbekannt"
	}

	reqURL := fmt.Sprintf("%s/api/states/%s", s.apiURL, sensorID)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return "unbekannt"
	}
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "unbekannt"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "unbekannt"
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "unbekannt"
	}

	var state stateResponse
	if err := json.Unmarshal(body, &state); err != nil {
		return "unbekannt"
	}

	if state.State == "" || state.State == "unavailable" || state.State == "unknown" {
		return "unbekannt"
	}

	if state.Attributes.UnitOfMeasurement != "" {
		return fmt.Sprintf("%s %s", state.State, state.Attributes.UnitOfMeasurement)
	}
	return state.State
}
