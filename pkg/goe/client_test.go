package goe

import (
	"encoding/json"
	"goe-report/pkg/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"
)

func TestClient_getApiTicket(t *testing.T) {
	// Start a local HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		if req.URL.Query().Get("filter") != "dll" {
			t.Errorf("Expected filter=dll, got %s", req.URL.Query().Get("filter"))
		}

		// Send response to be tested
		rw.Write([]byte(`{"dll": "https://data.v3.go-e.io/api/v1/direct_json?e=test-ticket"}`))
	}))
	// Close the server when test finishes
	defer server.Close()

	// Use Client instance with test server URL
	client := &Client{
		LocalApiUrl: server.URL,
		reqUrl:      server.URL,
	}

	ticket, err := client.getApiTicket()
	if err != nil {
		t.Fatalf("GetApiTicket failed: %v", err)
	}

	if ticket != "test-ticket" {
		t.Errorf("Expected ticket 'test-ticket', got '%s'", ticket)
	}
}

func TestClient_getApiTicket_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(`{"error": "something went wrong"}`))
	}))
	defer server.Close()

	client := &Client{
		LocalApiUrl: server.URL,
		reqUrl:      server.URL,
	}

	_, err := client.getApiTicket()
	// Should fail because JSON doesn't contain the "dll" field
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
}

func TestClient_getApiTicket_EmptyDll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(`{"dll": ""}`))
	}))
	defer server.Close()

	client := &Client{
		LocalApiUrl: server.URL,
		reqUrl:      server.URL,
	}

	_, err := client.getApiTicket()
	if err == nil {
		t.Fatalf("Expected error for empty dll, got nil")
	}
}

func TestClient_FetchChargingData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Query().Get("filter") == "dll" {
			rw.Write([]byte(`{"dll": "http://dummy?e=test-ticket"}`))
			return
		}

		// Verify URL parameters
		if req.URL.Query().Get("e") != "test-ticket" {
			t.Errorf("Expected e=test-ticket, got %s", req.URL.Query().Get("e"))
		}
		if req.URL.Query().Get("from") != "1600000000" {
			t.Errorf("Expected from=1600000000, got %s", req.URL.Query().Get("from"))
		}

		// Just returning dummy json
		rw.Write([]byte(`{
			"data": [
				{
					"id_chip": "12345",
					"id_chip_name": "TestChip",
					"start": "01.01.2023",
					"end": "01.01.2023",
					"seconds_total": "3600",
					"energy": 10.5
				}
			]
		}`))
	}))
	defer server.Close()

	client := &Client{
		reqUrl:        server.URL,
		directJsonUrl: server.URL,
	}

	resp, err := client.FetchChargingData(1600000000, 1600086400)
	if err != nil {
		t.Fatalf("FetchChargingData failed: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("Expected 1 data entry, got %d", len(resp.Data))
	}
	if resp.Data[0].IdChip != "12345" {
		t.Errorf("Expected IdChip 12345, got %v", resp.Data[0].IdChip)
	}
	if resp.Data[0].Energy != 10.5 {
		t.Errorf("Expected Energy 10.5, got %f", resp.Data[0].Energy)
	}
}

func TestClient_FetchChargingData_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Query().Get("filter") == "dll" {
			rw.Write([]byte(`{"dll": "http://dummy?e=test-ticket"}`))
			return
		}
		rw.Write([]byte(`{"invalid": json`)) // Invalid JSON to trigger parse error
	}))
	defer server.Close()

	client := &Client{
		reqUrl:        server.URL,
		directJsonUrl: server.URL,
	}

	_, err := client.FetchChargingData(0, 0)
	if err == nil {
		t.Fatalf("Expected error due to invalid JSON, got nil")
	}
}

func TestNewClient(t *testing.T) {
	// Setup config
	viper.Set(config.KeySerial, "123456")
	viper.Set(config.KeyToken, "test-token")
	viper.Set(config.KeyLocalApiUrl, "")
	defer viper.Reset()

	// Test Cloud instantiation
	client := NewClient()
	if client.reqUrl != "https://123456.api.v3.go-e.io/api/status?token=test-token" {
		t.Errorf("Unexpected cloud URL: %s", client.reqUrl)
	}

	// Test Local instantiation
	viper.Set(config.KeyLocalApiUrl, "http://192.168.1.100")
	clientLocal := NewClient()
	if clientLocal.reqUrl != "http://192.168.1.100/api/status" {
		t.Errorf("Unexpected local URL: %s", clientLocal.reqUrl)
	}
}

func TestClient_GetStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Mock the raw go-e API response
		mockData := rawStatusData{
			Car: 2,
			Alw: true,
			Amp: 16,
			Wh:  15000.0,
			Eto: 1500000.0,
			Nrg: []float64{230, 230, 230, 0, 16, 16, 16, 3680, 3680, 3680, 0, 11040, 0, 0, 0},
			Tma: []float64{35.5, 36.0},
			Frc: 0,
		}
		json.NewEncoder(rw).Encode(mockData)
	}))
	defer server.Close()

	client := &Client{
		reqUrl: server.URL,
	}

	status, err := client.GetStatus()
	if err != nil {
		t.Fatalf("GetStatus failed: %v", err)
	}

	if status.VehicleState != "Charging" {
		t.Errorf("Expected Charging, got %s", status.VehicleState)
	}
	if status.ChargingAllowed != "Yes" {
		t.Errorf("Expected Yes, got %s", status.ChargingAllowed)
	}
	if status.SetCurrentA != 16 {
		t.Errorf("Expected 16, got %d", status.SetCurrentA)
	}
	if status.ChargedSincePlugInKWh != 15.0 { // 15000 / 1000
		t.Errorf("Expected 15.0, got %f", status.ChargedSincePlugInKWh)
	}
	if status.TotalEnergyLifetimeKWh != 1500.0 { // 1500000 / 1000
		t.Errorf("Expected 1500.0, got %f", status.TotalEnergyLifetimeKWh)
	}
	if status.CurrentPowerKW != 11.04 { // 11040 / 1000
		t.Errorf("Expected 11.04, got %f", status.CurrentPowerKW)
	}
	if status.TemperatureCelsius != "35.5 °C" {
		t.Errorf("Expected 35.5 °C, got %s", status.TemperatureCelsius)
	}
	if len(status.Phases) != 3 {
		t.Errorf("Expected 3 phases, got %d", len(status.Phases))
	} else {
		p1 := status.Phases[0]
		if p1.Voltage != 230 || p1.Current != 16 || p1.Power != 3680 {
			t.Errorf("Phase 1 incorrect: %+v", p1)
		}
	}
}

func TestRawStatusData_toDTO(t *testing.T) {
	// Test edge cases in mapping that we might missed
	raw := &rawStatusData{
		Car: 99, // Unknown state
		Alw: false,
		Nrg: []float64{230, 230}, // Short array
		Tma: []float64{},         // Empty short array
	}

	dto := raw.toDTO()

	if dto.VehicleState != "Unknown" {
		t.Errorf("Expected Unknown for car=99, got %s", dto.VehicleState)
	}
	if dto.ChargingAllowed != "No" {
		t.Errorf("Expected No for alw=false, got %s", dto.ChargingAllowed)
	}
	if dto.CurrentPowerKW != 0 {
		t.Errorf("Expected 0.0 for short nrg array, got %f", dto.CurrentPowerKW)
	}
	if dto.TemperatureCelsius != "N/A" {
		t.Errorf("Expected N/A for empty tma array, got %s", dto.TemperatureCelsius)
	}
	if len(dto.Phases) != 0 {
		t.Errorf("Expected 0 phases for short nrg array, got %d", len(dto.Phases))
	}
}
