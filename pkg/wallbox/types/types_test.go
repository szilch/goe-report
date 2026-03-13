package types

import (
	"testing"
)

func TestPhaseDetail_Struct(t *testing.T) {
	phase := PhaseDetail{
		Voltage: 230.5,
		Current: 16.0,
		Power:   3688.0,
	}

	if phase.Voltage != 230.5 {
		t.Errorf("Expected Voltage 230.5, got: %f", phase.Voltage)
	}
	if phase.Current != 16.0 {
		t.Errorf("Expected Current 16.0, got: %f", phase.Current)
	}
	if phase.Power != 3688.0 {
		t.Errorf("Expected Power 3688.0, got: %f", phase.Power)
	}
}

func TestPhaseDetail_ZeroValues(t *testing.T) {
	phase := PhaseDetail{}

	if phase.Voltage != 0 {
		t.Errorf("Expected Voltage 0, got: %f", phase.Voltage)
	}
	if phase.Current != 0 {
		t.Errorf("Expected Current 0, got: %f", phase.Current)
	}
	if phase.Power != 0 {
		t.Errorf("Expected Power 0, got: %f", phase.Power)
	}
}

func TestStatus_Struct(t *testing.T) {
	status := Status{
		VehicleState:           "Charging",
		ChargingAllowed:        "Yes",
		SetCurrentA:            16,
		CurrentPowerKW:         11.04,
		ChargedSincePlugInKWh:  25.5,
		TotalEnergyLifetimeKWh: 1500.75,
		TemperatureCelsius:     "32.5 °C",
		Phases: []PhaseDetail{
			{Voltage: 230.0, Current: 16.0, Power: 3680.0},
			{Voltage: 231.0, Current: 16.0, Power: 3696.0},
			{Voltage: 229.0, Current: 16.0, Power: 3664.0},
		},
	}

	if status.VehicleState != "Charging" {
		t.Errorf("Expected VehicleState 'Charging', got: %s", status.VehicleState)
	}
	if status.ChargingAllowed != "Yes" {
		t.Errorf("Expected ChargingAllowed 'Yes', got: %s", status.ChargingAllowed)
	}
	if status.SetCurrentA != 16 {
		t.Errorf("Expected SetCurrentA 16, got: %d", status.SetCurrentA)
	}
	if status.CurrentPowerKW != 11.04 {
		t.Errorf("Expected CurrentPowerKW 11.04, got: %f", status.CurrentPowerKW)
	}
	if status.ChargedSincePlugInKWh != 25.5 {
		t.Errorf("Expected ChargedSincePlugInKWh 25.5, got: %f", status.ChargedSincePlugInKWh)
	}
	if status.TotalEnergyLifetimeKWh != 1500.75 {
		t.Errorf("Expected TotalEnergyLifetimeKWh 1500.75, got: %f", status.TotalEnergyLifetimeKWh)
	}
	if status.TemperatureCelsius != "32.5 °C" {
		t.Errorf("Expected TemperatureCelsius '32.5 °C', got: %s", status.TemperatureCelsius)
	}
	if len(status.Phases) != 3 {
		t.Errorf("Expected 3 Phases, got: %d", len(status.Phases))
	}
}

func TestStatus_EmptyPhases(t *testing.T) {
	status := Status{
		VehicleState: "Idle",
		Phases:       nil,
	}

	if status.Phases != nil {
		t.Error("Expected Phases to be nil")
	}
}

func TestChargingSession_Struct(t *testing.T) {
	session := ChargingSession{
		IdChip:       "chip123",
		IdChipName:   "My RFID Card",
		Start:        "2024-01-15 08:00:00",
		End:          "2024-01-15 12:00:00",
		SecondsTotal: "14400",
		Energy:       45.75,
	}

	if session.IdChip != "chip123" {
		t.Errorf("Expected IdChip 'chip123', got: %v", session.IdChip)
	}
	if session.IdChipName != "My RFID Card" {
		t.Errorf("Expected IdChipName 'My RFID Card', got: %s", session.IdChipName)
	}
	if session.Start != "2024-01-15 08:00:00" {
		t.Errorf("Expected Start '2024-01-15 08:00:00', got: %s", session.Start)
	}
	if session.End != "2024-01-15 12:00:00" {
		t.Errorf("Expected End '2024-01-15 12:00:00', got: %s", session.End)
	}
	if session.SecondsTotal != "14400" {
		t.Errorf("Expected SecondsTotal '14400', got: %s", session.SecondsTotal)
	}
	if session.Energy != 45.75 {
		t.Errorf("Expected Energy 45.75, got: %f", session.Energy)
	}
}

func TestChargingSession_IdChipAsInt(t *testing.T) {
	// IdChip can be different types (interface{})
	session := ChargingSession{
		IdChip: 12345,
	}

	if session.IdChip != 12345 {
		t.Errorf("Expected IdChip 12345, got: %v", session.IdChip)
	}
}

func TestChargingSession_IdChipAsNil(t *testing.T) {
	session := ChargingSession{
		IdChip: nil,
	}

	if session.IdChip != nil {
		t.Errorf("Expected IdChip nil, got: %v", session.IdChip)
	}
}

func TestChargingResponse_Struct(t *testing.T) {
	response := ChargingResponse{
		Data: []ChargingSession{
			{IdChipName: "Card 1", Energy: 10.0},
			{IdChipName: "Card 2", Energy: 20.0},
		},
	}

	if len(response.Data) != 2 {
		t.Errorf("Expected 2 sessions, got: %d", len(response.Data))
	}
	if response.Data[0].IdChipName != "Card 1" {
		t.Errorf("Expected first session IdChipName 'Card 1', got: %s", response.Data[0].IdChipName)
	}
	if response.Data[1].Energy != 20.0 {
		t.Errorf("Expected second session Energy 20.0, got: %f", response.Data[1].Energy)
	}
}

func TestChargingResponse_Empty(t *testing.T) {
	response := ChargingResponse{
		Data: []ChargingSession{},
	}

	if len(response.Data) != 0 {
		t.Errorf("Expected 0 sessions, got: %d", len(response.Data))
	}
}

func TestChargingResponse_Nil(t *testing.T) {
	response := ChargingResponse{}

	if response.Data != nil {
		t.Error("Expected Data to be nil")
	}
}

// Test that interface Adapter has the expected methods
// This is a compile-time check
func TestAdapter_Interface(t *testing.T) {
	// This is just a compile-time verification that the interface exists
	// and has the expected method signatures
	var _ Adapter = nil
}

// MockAdapter for testing purposes
type MockAdapter struct {
	FetchChargingDataFunc func(fromMs, toMs int64) (*ChargingResponse, error)
	GetStatusFunc         func() (*Status, error)
	GetTypeFunc           func() string
}

func (m *MockAdapter) FetchChargingData(fromMs, toMs int64) (*ChargingResponse, error) {
	if m.FetchChargingDataFunc != nil {
		return m.FetchChargingDataFunc(fromMs, toMs)
	}
	return &ChargingResponse{Data: []ChargingSession{}}, nil
}

func (m *MockAdapter) GetStatus() (*Status, error) {
	if m.GetStatusFunc != nil {
		return m.GetStatusFunc()
	}
	return &Status{}, nil
}

func (m *MockAdapter) GetType() string {
	if m.GetTypeFunc != nil {
		return m.GetTypeFunc()
	}
	return "mock"
}

func TestMockAdapter_ImplementsInterface(t *testing.T) {
	// Verify MockAdapter implements Adapter interface
	var adapter Adapter = &MockAdapter{}

	if adapter == nil {
		t.Error("MockAdapter should not be nil")
	}
}

func TestMockAdapter_FetchChargingData(t *testing.T) {
	expectedResponse := &ChargingResponse{
		Data: []ChargingSession{
			{IdChipName: "Test", Energy: 100.0},
		},
	}

	mock := &MockAdapter{
		FetchChargingDataFunc: func(fromMs, toMs int64) (*ChargingResponse, error) {
			return expectedResponse, nil
		},
	}

	response, err := mock.FetchChargingData(1000, 2000)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if response != expectedResponse {
		t.Error("Response does not match expected")
	}
}

func TestMockAdapter_GetStatus(t *testing.T) {
	expectedStatus := &Status{
		VehicleState: "Charging",
		SetCurrentA:  16,
	}

	mock := &MockAdapter{
		GetStatusFunc: func() (*Status, error) {
			return expectedStatus, nil
		},
	}

	status, err := mock.GetStatus()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if status != expectedStatus {
		t.Error("Status does not match expected")
	}
}

func TestMockAdapter_GetType(t *testing.T) {
	mock := &MockAdapter{
		GetTypeFunc: func() string {
			return "custom-type"
		},
	}

	adapterType := mock.GetType()

	if adapterType != "custom-type" {
		t.Errorf("Expected type 'custom-type', got: %s", adapterType)
	}
}
