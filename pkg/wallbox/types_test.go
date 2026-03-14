package wallbox

import (
	"echarge-report/pkg/wallbox/types"
	"testing"
)

// Test that type aliases work correctly
func TestTypeAliases_PhaseDetail(t *testing.T) {
	phase := PhaseDetail{
		Voltage: 230.0,
		Current: 16.0,
		Power:   3680.0,
	}
	var originalType types.PhaseDetail = phase
	if originalType.Voltage != 230.0 {
		t.Errorf("Expected Voltage 230.0, got: %f", originalType.Voltage)
	}
}
func TestTypeAliases_Status(t *testing.T) {
	status := Status{
		VehicleState:           "Charging",
		ChargingAllowed:        "Yes",
		SetCurrentA:            16,
		CurrentPowerKW:         11.04,
		ChargedSincePlugInKWh:  25.5,
		TotalEnergyLifetimeKWh: 1500.0,
		TemperatureCelsius:     "32.5 °C",
		Phases: []PhaseDetail{
			{Voltage: 230.0, Current: 16.0, Power: 3680.0},
		},
	}
	var originalType types.Status = status
	if originalType.VehicleState != "Charging" {
		t.Errorf("Expected VehicleState 'Charging', got: %s", originalType.VehicleState)
	}
	if len(originalType.Phases) != 1 {
		t.Errorf("Expected 1 phase, got: %d", len(originalType.Phases))
	}
}
func TestTypeAliases_ChargingSession(t *testing.T) {
	session := ChargingSession{
		IdChip:       "chip123",
		IdChipName:   "Test Card",
		Start:        "2024-01-01 10:00:00",
		End:          "2024-01-01 12:00:00",
		SecondsTotal: "7200",
		Energy:       15.5,
	}
	var originalType types.ChargingSession = session
	if originalType.IdChipName != "Test Card" {
		t.Errorf("Expected IdChipName 'Test Card', got: %s", originalType.IdChipName)
	}
	if originalType.Energy != 15.5 {
		t.Errorf("Expected Energy 15.5, got: %f", originalType.Energy)
	}
}
func TestTypeAliases_ChargingResponse(t *testing.T) {
	response := ChargingResponse{
		Data: []ChargingSession{
			{IdChipName: "Card 1", Energy: 10.0},
			{IdChipName: "Card 2", Energy: 20.0},
		},
	}
	var originalType types.ChargingResponse = response
	if len(originalType.Data) != 2 {
		t.Errorf("Expected 2 sessions, got: %d", len(originalType.Data))
	}
}
func TestTypeAliases_Adapter(t *testing.T) {
	var adapter Adapter = nil
	var originalAdapter types.Adapter = adapter
	if originalAdapter != nil {
		t.Error("Expected nil adapter")
	}
}
func TestTypeAliases_Interoperability(t *testing.T) {
	aliasPhase := PhaseDetail{Voltage: 230.0, Current: 16.0, Power: 3680.0}
	originalPhase := types.PhaseDetail{Voltage: 231.0, Current: 15.0, Power: 3465.0}
	phases := []PhaseDetail{aliasPhase, originalPhase}
	if len(phases) != 2 {
		t.Errorf("Expected 2 phases, got: %d", len(phases))
	}
	if phases[0].Voltage != 230.0 {
		t.Errorf("Expected first phase voltage 230.0, got: %f", phases[0].Voltage)
	}
	if phases[1].Voltage != 231.0 {
		t.Errorf("Expected second phase voltage 231.0, got: %f", phases[1].Voltage)
	}
}
