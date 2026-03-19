package wallbox

import (
	"echarge-report/pkg/config"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestNewGoeAdapter(t *testing.T) {
	viper.Set(config.KeyWallboxGoeCloudSerial, "test-serial-123")
	viper.Set(config.KeyWallboxGoeCloudToken, "test-token-456")
	viper.Set(config.KeyWallboxGoeLocalApiUrl, "")
	defer viper.Reset()

	adapter := newGoeAdapter()

	if adapter == nil {
		t.Fatal("newGoeAdapter() returned nil")
	}
	if adapter.serial != "test-serial-123" {
		t.Errorf("Expected serial 'test-serial-123', got: %s", adapter.serial)
	}
	if adapter.token != "test-token-456" {
		t.Errorf("Expected token 'test-token-456', got: %s", adapter.token)
	}
}

func TestNewGoeAdapter_WithLocalApiUrl(t *testing.T) {
	viper.Set(config.KeyWallboxGoeCloudSerial, "test-serial")
	viper.Set(config.KeyWallboxGoeCloudToken, "test-token")
	viper.Set(config.KeyWallboxGoeLocalApiUrl, "http://192.168.1.100")
	defer viper.Reset()

	adapter := newGoeAdapter()

	if adapter == nil {
		t.Fatal("newGoeAdapter() returned nil")
	}
	if adapter.localAPIURL != "http://192.168.1.100" {
		t.Errorf("Expected localAPIURL 'http://192.168.1.100', got: %s", adapter.localAPIURL)
	}
	expectedReqURL := "http://192.168.1.100/api/status"
	if adapter.reqURL != expectedReqURL {
		t.Errorf("Expected reqURL '%s', got: %s", expectedReqURL, adapter.reqURL)
	}
}

func TestNewGoeAdapter_WithCloudApi(t *testing.T) {
	viper.Set(config.KeyWallboxGoeCloudSerial, "ABC123")
	viper.Set(config.KeyWallboxGoeCloudToken, "secret-token")
	viper.Set(config.KeyWallboxGoeLocalApiUrl, "")
	defer viper.Reset()

	adapter := newGoeAdapter()

	if adapter == nil {
		t.Fatal("newGoeAdapter() returned nil")
	}
	expectedReqURL := "https://ABC123.api.v3.go-e.io/api/status?token=secret-token"
	if adapter.reqURL != expectedReqURL {
		t.Errorf("Expected reqURL '%s', got: %s", expectedReqURL, adapter.reqURL)
	}
}

func TestGoeAdapter_GetType(t *testing.T) {
	adapter := &goeAdapter{}

	adapterType := adapter.GetType()

	if adapterType != "goe" {
		t.Errorf("Expected type 'goe', got: %s", adapterType)
	}
}

func TestGoeAdapter_ImplementsInterface(t *testing.T) {
	var _ Adapter = (*goeAdapter)(nil)
}

func TestGoeAdapter_ToStatus_Charging(t *testing.T) {
	adapter := &goeAdapter{}

	raw := &rawStatusData{
		Car: 2,
		Alw: true,
		Amp: 16,
		Wh:  5000.0,
		Eto: 10000.0,
		Nrg: []float64{230.0, 231.0, 229.0, 0, 10.0, 11.0, 9.0, 2300.0, 2541.0, 2061.0, 0, 6902.0},
		Tma: []float64{25.5, 24.0},
	}

	status := adapter.toStatus(raw)

	if status.VehicleState != "Charging" {
		t.Errorf("Expected VehicleState 'Charging', got: %s", status.VehicleState)
	}
	if !status.ChargingAllowed {
		t.Errorf("Expected ChargingAllowed true, got: %v", status.ChargingAllowed)
	}
	if status.SetCurrentA != 16 {
		t.Errorf("Expected SetCurrentA 16, got: %d", status.SetCurrentA)
	}
	if status.ChargedSincePlugInKWh != 5.0 {
		t.Errorf("Expected ChargedSincePlugInKWh 5.0, got: %f", status.ChargedSincePlugInKWh)
	}
	if status.TotalEnergyLifetimeKWh != 10.0 {
		t.Errorf("Expected TotalEnergyLifetimeKWh 10.0, got: %f", status.TotalEnergyLifetimeKWh)
	}
	if status.TemperatureCelsius != "25.5 °C" {
		t.Errorf("Expected TemperatureCelsius '25.5 °C', got: %s", status.TemperatureCelsius)
	}
	if status.CurrentPowerKW != 6.902 {
		t.Errorf("Expected CurrentPowerKW 6.902, got: %f", status.CurrentPowerKW)
	}
}

func TestGoeAdapter_ToStatus_AllVehicleStates(t *testing.T) {
	adapter := &goeAdapter{}

	testCases := []struct {
		carState     int
		expectedText string
	}{
		{1, "Idle (not connected)"},
		{2, "Charging"},
		{3, "Waiting for car"},
		{4, "Charging complete"},
		{5, "Error"},
		{99, "Unknown"},
	}

	for _, tc := range testCases {
		raw := &rawStatusData{Car: tc.carState, Nrg: []float64{}, Tma: []float64{}}
		status := adapter.toStatus(raw)

		if status.VehicleState != tc.expectedText {
			t.Errorf("For car state %d: expected VehicleState '%s', got: %s",
				tc.carState, tc.expectedText, status.VehicleState)
		}
	}
}

func TestGoeAdapter_ToStatus_ChargingNotAllowed(t *testing.T) {
	adapter := &goeAdapter{}

	raw := &rawStatusData{
		Car: 1,
		Alw: false,
		Nrg: []float64{},
		Tma: []float64{},
	}

	status := adapter.toStatus(raw)

	if status.ChargingAllowed {
		t.Errorf("Expected ChargingAllowed false, got: %v", status.ChargingAllowed)
	}
}

func TestGoeAdapter_ToStatus_NoTemperatureData(t *testing.T) {
	adapter := &goeAdapter{}

	raw := &rawStatusData{
		Car: 1,
		Nrg: []float64{},
		Tma: []float64{},
	}

	status := adapter.toStatus(raw)

	if status.TemperatureCelsius != "N/A" {
		t.Errorf("Expected TemperatureCelsius 'N/A', got: %s", status.TemperatureCelsius)
	}
}

func TestGoeAdapter_ToStatus_PhaseDetails(t *testing.T) {
	adapter := &goeAdapter{}

	raw := &rawStatusData{
		Car: 2,
		Nrg: []float64{230.0, 231.0, 229.0, 0, 10.0, 11.0, 9.0, 2300.0, 2541.0, 2061.0, 0, 6902.0},
		Tma: []float64{},
	}

	status := adapter.toStatus(raw)

	if len(status.Phases) != 3 {
		t.Fatalf("Expected 3 phases, got: %d", len(status.Phases))
	}
	if status.Phases[0].Voltage != 230.0 {
		t.Errorf("Phase 1 voltage: expected 230.0, got: %f", status.Phases[0].Voltage)
	}
	if status.Phases[0].Current != 10.0 {
		t.Errorf("Phase 1 current: expected 10.0, got: %f", status.Phases[0].Current)
	}
	if status.Phases[0].Power != 2300.0 {
		t.Errorf("Phase 1 power: expected 2300.0, got: %f", status.Phases[0].Power)
	}
	if status.Phases[1].Voltage != 231.0 {
		t.Errorf("Phase 2 voltage: expected 231.0, got: %f", status.Phases[1].Voltage)
	}
	if status.Phases[2].Voltage != 229.0 {
		t.Errorf("Phase 3 voltage: expected 229.0, got: %f", status.Phases[2].Voltage)
	}
}

func TestGoeAdapter_ToStatus_InsufficientNrgData(t *testing.T) {
	adapter := &goeAdapter{}

	raw := &rawStatusData{
		Car: 1,
		Nrg: []float64{230.0, 231.0},
		Tma: []float64{},
	}

	status := adapter.toStatus(raw)

	if len(status.Phases) != 0 {
		t.Errorf("Expected 0 phases when insufficient data, got: %d", len(status.Phases))
	}
	if status.CurrentPowerKW != 0 {
		t.Errorf("Expected CurrentPowerKW 0 when insufficient data, got: %f", status.CurrentPowerKW)
	}
}

func TestGoeAdapter_ToChargingResponse(t *testing.T) {
	adapter := &goeAdapter{}

	rawData := &directJsonResp{
		Data: []chargingLogRaw{
			{
				IdChip:       "chip1",
				IdChipName:   "Card 1",
				Start:        "01.01.2024 10:00:00",
				End:          "01.01.2024 12:00:00",
				SecondsTotal: "02:00:00",
				Energy:       15.5,
			},
			{
				IdChip:       "chip2",
				IdChipName:   "Card 2",
				Start:        "02.01.2024 14:00:00",
				End:          "02.01.2024 16:30:00",
				SecondsTotal: "02:30:00",
				Energy:       22.3,
			},
		},
	}

	response := adapter.toChargingResponse(rawData)

	if response == nil {
		t.Fatal("toChargingResponse returned nil")
	}
	if len(response.Data) != 2 {
		t.Fatalf("Expected 2 sessions, got: %d", len(response.Data))
	}
	if response.Data[0].IdChipName != "Card 1" {
		t.Errorf("Expected IdChipName 'Card 1', got: %s", response.Data[0].IdChipName)
	}
	if response.Data[0].Energy != 15.5 {
		t.Errorf("Expected Energy 15.5, got: %f", response.Data[0].Energy)
	}
	locBerlin, _ := time.LoadLocation("Europe/Berlin")
	expectedStart1, _ := time.ParseInLocation("02.01.2006 15:04:05", "01.01.2024 11:00:00", locBerlin)
	if !response.Data[0].Start.Equal(expectedStart1) {
		t.Errorf("Expected Start1 %v, got: %v", expectedStart1, response.Data[0].Start)
	}
	
	if response.Data[0].Duration != 2*time.Hour {
		t.Errorf("Expected Duration1 %v, got: %v", 2*time.Hour, response.Data[0].Duration)
	}
	
	expectedStart2, _ := time.ParseInLocation("02.01.2006 15:04:05", "02.01.2024 15:00:00", locBerlin)
	if !response.Data[1].Start.Equal(expectedStart2) {
		t.Errorf("Expected Start2 %v, got: %v", expectedStart2, response.Data[1].Start)
	}
	if response.Data[1].Duration != 2*time.Hour+30*time.Minute {
		t.Errorf("Expected Duration2 %v, got: %v", 2*time.Hour+30*time.Minute, response.Data[1].Duration)
	}
	if response.Data[1].Energy != 22.3 {
		t.Errorf("Expected Energy 22.3, got: %f", response.Data[1].Energy)
	}
}

func TestGoeAdapter_ToChargingResponse_Empty(t *testing.T) {
	adapter := &goeAdapter{}

	rawData := &directJsonResp{
		Data: []chargingLogRaw{},
	}

	response := adapter.toChargingResponse(rawData)

	if response == nil {
		t.Fatal("toChargingResponse returned nil")
	}
	if len(response.Data) != 0 {
		t.Errorf("Expected 0 sessions, got: %d", len(response.Data))
	}
}

func TestGoeAdapter_GetStatus_Success(t *testing.T) {
	statusResponse := rawStatusData{
		Car: 2,
		Alw: true,
		Amp: 16,
		Wh:  5000.0,
		Eto: 10000.0,
		Nrg: []float64{230.0, 231.0, 229.0, 0, 10.0, 11.0, 9.0, 2300.0, 2541.0, 2061.0, 0, 6902.0},
		Tma: []float64{25.5},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(statusResponse)
	}))
	defer server.Close()

	adapter := &goeAdapter{
		reqURL: server.URL,
	}

	status, err := adapter.GetStatus()

	if err != nil {
		t.Fatalf("GetStatus() returned error: %v", err)
	}
	if status == nil {
		t.Fatal("GetStatus() returned nil status")
	}
	if status.VehicleState != "Charging" {
		t.Errorf("Expected VehicleState 'Charging', got: %s", status.VehicleState)
	}
	if status.SetCurrentA != 16 {
		t.Errorf("Expected SetCurrentA 16, got: %d", status.SetCurrentA)
	}
}

func TestGoeAdapter_GetStatus_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	adapter := &goeAdapter{
		reqURL: server.URL,
	}

	status, err := adapter.GetStatus()

	if err == nil {
		t.Error("GetStatus() should return error for HTTP 500")
	}
	if status != nil {
		t.Error("GetStatus() should return nil status on error")
	}
}

func TestGoeAdapter_GetStatus_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	adapter := &goeAdapter{
		reqURL: server.URL,
	}

	status, err := adapter.GetStatus()

	if err == nil {
		t.Error("GetStatus() should return error for invalid JSON")
	}
	if status != nil {
		t.Error("GetStatus() should return nil status on error")
	}
}

func TestGoeAdapter_GetStatus_ConnectionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	server.Close()

	adapter := &goeAdapter{
		reqURL: server.URL,
	}

	status, err := adapter.GetStatus()

	if err == nil {
		t.Error("GetStatus() should return error for connection failure")
	}
	if status != nil {
		t.Error("GetStatus() should return nil status on error")
	}
}

func TestGoeChargingLogRaw_JsonUnmarshal(t *testing.T) {
	jsonData := `{
		"id_chip": "ABC123",
		"id_chip_name": "My Card",
		"start": "2024-01-15 08:00:00",
		"end": "2024-01-15 10:30:00",
		"seconds_total": "02:30:00",
		"energy": 25.75
	}`

	var log chargingLogRaw
	err := json.Unmarshal([]byte(jsonData), &log)

	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	if log.IdChip != "ABC123" {
		t.Errorf("Expected IdChip 'ABC123', got: %v", log.IdChip)
	}
	if log.IdChipName != "My Card" {
		t.Errorf("Expected IdChipName 'My Card', got: %s", log.IdChipName)
	}
	if log.Energy != 25.75 {
		t.Errorf("Expected Energy 25.75, got: %f", log.Energy)
	}
}

func TestGoeDirectJsonResp_JsonUnmarshal(t *testing.T) {
	jsonData := `{
		"data": [
			{
				"id_chip": "chip1",
				"id_chip_name": "Card 1",
				"start": "2024-01-01 10:00:00",
				"end": "2024-01-01 12:00:00",
				"seconds_total": "02:00:00",
				"energy": 15.5
			}
		]
	}`

	var resp directJsonResp
	err := json.Unmarshal([]byte(jsonData), &resp)

	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("Expected 1 data entry, got: %d", len(resp.Data))
	}
	if resp.Data[0].IdChipName != "Card 1" {
		t.Errorf("Expected IdChipName 'Card 1', got: %s", resp.Data[0].IdChipName)
	}
}

func TestGoeRawStatusData_JsonUnmarshal(t *testing.T) {
	jsonData := `{
		"car": 2,
		"alw": true,
		"amp": 16,
		"wh": 5000.0,
		"eto": 10000.0,
		"nrg": [230.0, 231.0, 229.0, 0, 10.0, 11.0, 9.0, 2300.0, 2541.0, 2061.0, 0, 6902.0],
		"tma": [25.5, 24.0],
		"frc": 0
	}`

	var raw rawStatusData
	err := json.Unmarshal([]byte(jsonData), &raw)

	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}
	if raw.Car != 2 {
		t.Errorf("Expected Car 2, got: %d", raw.Car)
	}
	if !raw.Alw {
		t.Error("Expected Alw true, got: false")
	}
	if raw.Amp != 16 {
		t.Errorf("Expected Amp 16, got: %d", raw.Amp)
	}
	if len(raw.Nrg) != 12 {
		t.Errorf("Expected 12 Nrg values, got: %d", len(raw.Nrg))
	}
	if len(raw.Tma) != 2 {
		t.Errorf("Expected 2 Tma values, got: %d", len(raw.Tma))
	}
}

func TestGoeAdapter_ParseDuration_Invalid(t *testing.T) {
	adapter := &goeAdapter{}

	d := adapter.parseDuration("not-a-duration")

	if d != 0 {
		t.Errorf("Expected 0 duration for invalid input, got: %v", d)
	}
}

func TestGoeAdapter_FetchChargingData_Success(t *testing.T) {
	ticket := "test-ticket-123"
	
	// Server to mock go-e API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("filter") == "dll" {
			// Step 1: getAPITicket
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"dll": "https://data.v3.go-e.io/api/v1/direct_json?e=%s"}`, ticket)
			return
		}
		
		if r.URL.Query().Get("e") == ticket {
			// Step 2: FetchChargingData
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{
				"data": [
					{
						"id_chip": "chip1",
						"id_chip_name": "Card 1",
						"start": "01.01.2024 10:00:00",
						"end": "01.01.2024 12:00:00",
						"seconds_total": "02:00:00",
						"energy": 15.5
					}
				]
			}`)
			return
		}
		
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	adapter := &goeAdapter{
		reqURL:        server.URL,
		directJSONURL: server.URL,
	}

	resp, err := adapter.FetchChargingData(0, 0)

	if err != nil {
		t.Fatalf("FetchChargingData() returned error: %v", err)
	}
	if resp == nil || len(resp.Data) != 1 {
		t.Fatalf("Expected 1 session, got: %v", resp)
	}
	if resp.Data[0].IdChipName != "Card 1" {
		t.Errorf("Expected IdChipName 'Card 1', got: %s", resp.Data[0].IdChipName)
	}
}

func TestGoeAdapter_FetchChargingData_TicketError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"dll": ""}`) // Empty ticket
	}))
	defer server.Close()

	adapter := &goeAdapter{
		reqURL: server.URL,
	}

	_, err := adapter.FetchChargingData(0, 0)

	if err == nil {
		t.Error("Expected error when API returns no ticket, got nil")
	}
}

func TestGoeAdapter_FetchChargingData_HTTPError(t *testing.T) {
	// First call (ticket) succeeds, second call (data) fails
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"dll": "https://example.com/api?e=ticket"}`)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	adapter := &goeAdapter{
		reqURL:        server.URL,
		directJSONURL: server.URL,
	}

	_, err := adapter.FetchChargingData(0, 0)

	if err == nil {
		t.Error("Expected error when data fetch fails, got nil")
	}
}

func TestGoeAdapter_ParseTime_Invalid(t *testing.T) {
	adapter := &goeAdapter{}
	
	t1 := adapter.parseTime("invalid-time")
	if !t1.IsZero() {
		t.Errorf("Expected zero time for invalid input, got %v", t1)
	}
}
