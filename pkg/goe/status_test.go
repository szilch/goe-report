package goe

import (
	"testing"
)

func TestRawStatusData_ToDTO_VehicleStates(t *testing.T) {
	tests := []struct {
		name     string
		carState int
		expected string
	}{
		{"Idle", 1, "Idle (not connected)"},
		{"Charging", 2, "Charging"},
		{"Waiting", 3, "Waiting for car"},
		{"Complete", 4, "Charging complete"},
		{"Error", 5, "Error"},
		{"Unknown", 99, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := &rawStatusData{
				Car: tt.carState,
				Nrg: make([]float64, 12),
			}
			dto := raw.toDTO()

			if dto.VehicleState != tt.expected {
				t.Errorf("expected VehicleState '%s', got '%s'", tt.expected, dto.VehicleState)
			}
		})
	}
}

func TestRawStatusData_ToDTO_ChargingAllowed(t *testing.T) {
	tests := []struct {
		name     string
		alw      bool
		expected string
	}{
		{"Allowed", true, "Yes"},
		{"NotAllowed", false, "No"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := &rawStatusData{
				Alw: tt.alw,
				Nrg: make([]float64, 12),
			}
			dto := raw.toDTO()

			if dto.ChargingAllowed != tt.expected {
				t.Errorf("expected ChargingAllowed '%s', got '%s'", tt.expected, dto.ChargingAllowed)
			}
		})
	}
}

func TestRawStatusData_ToDTO_EnergyConversion(t *testing.T) {
	raw := &rawStatusData{
		Wh:  15000,  // 15 kWh
		Eto: 500000, // 500 kWh
		Nrg: make([]float64, 12),
	}
	dto := raw.toDTO()

	expectedPlugIn := 15.0
	if dto.ChargedSincePlugInKWh != expectedPlugIn {
		t.Errorf("expected ChargedSincePlugInKWh %.2f, got %.2f", expectedPlugIn, dto.ChargedSincePlugInKWh)
	}

	expectedLifetime := 500.0
	if dto.TotalEnergyLifetimeKWh != expectedLifetime {
		t.Errorf("expected TotalEnergyLifetimeKWh %.2f, got %.2f", expectedLifetime, dto.TotalEnergyLifetimeKWh)
	}
}

func TestRawStatusData_ToDTO_SetCurrentA(t *testing.T) {
	raw := &rawStatusData{
		Amp: 16,
		Nrg: make([]float64, 12),
	}
	dto := raw.toDTO()

	if dto.SetCurrentA != 16 {
		t.Errorf("expected SetCurrentA 16, got %d", dto.SetCurrentA)
	}
}

func TestRawStatusData_ToDTO_CurrentPower(t *testing.T) {
	raw := &rawStatusData{
		Nrg: []float64{230, 231, 229, 0, 10, 11, 9, 2300, 2541, 2061, 0, 6900},
	}
	dto := raw.toDTO()

	expectedPower := 6.9
	if dto.CurrentPowerKW != expectedPower {
		t.Errorf("expected CurrentPowerKW %.2f, got %.2f", expectedPower, dto.CurrentPowerKW)
	}
}

func TestRawStatusData_ToDTO_Temperature(t *testing.T) {
	tests := []struct {
		name     string
		tma      []float64
		expected string
	}{
		{"WithTemperature", []float64{35.5, 40.2}, "35.5 °C"},
		{"EmptyTma", []float64{}, "N/A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw := &rawStatusData{
				Tma: tt.tma,
				Nrg: make([]float64, 12),
			}
			dto := raw.toDTO()

			if dto.TemperatureCelsius != tt.expected {
				t.Errorf("expected TemperatureCelsius '%s', got '%s'", tt.expected, dto.TemperatureCelsius)
			}
		})
	}
}

func TestRawStatusData_ToDTO_Phases(t *testing.T) {
	raw := &rawStatusData{
		Nrg: []float64{230.5, 231.2, 229.8, 0, 10.1, 11.2, 9.8, 2321.55, 2589.44, 2250.84, 0, 7161.83},
	}
	dto := raw.toDTO()

	if len(dto.Phases) != 3 {
		t.Fatalf("expected 3 phases, got %d", len(dto.Phases))
	}

	// Phase L1
	if dto.Phases[0].Voltage != 230.5 {
		t.Errorf("expected L1 Voltage 230.5, got %.1f", dto.Phases[0].Voltage)
	}
	if dto.Phases[0].Current != 10.1 {
		t.Errorf("expected L1 Current 10.1, got %.1f", dto.Phases[0].Current)
	}
	if dto.Phases[0].Power != 2321.55 {
		t.Errorf("expected L1 Power 2321.55, got %.2f", dto.Phases[0].Power)
	}

	// Phase L2
	if dto.Phases[1].Voltage != 231.2 {
		t.Errorf("expected L2 Voltage 231.2, got %.1f", dto.Phases[1].Voltage)
	}
	if dto.Phases[1].Current != 11.2 {
		t.Errorf("expected L2 Current 11.2, got %.1f", dto.Phases[1].Current)
	}
	if dto.Phases[1].Power != 2589.44 {
		t.Errorf("expected L2 Power 2589.44, got %.2f", dto.Phases[1].Power)
	}

	// Phase L3
	if dto.Phases[2].Voltage != 229.8 {
		t.Errorf("expected L3 Voltage 229.8, got %.1f", dto.Phases[2].Voltage)
	}
	if dto.Phases[2].Current != 9.8 {
		t.Errorf("expected L3 Current 9.8, got %.1f", dto.Phases[2].Current)
	}
	if dto.Phases[2].Power != 2250.84 {
		t.Errorf("expected L3 Power 2250.84, got %.2f", dto.Phases[2].Power)
	}
}

func TestRawStatusData_ToDTO_InsufficientNrgArray(t *testing.T) {
	raw := &rawStatusData{
		Nrg: []float64{230, 231, 229}, // Only 3 elements
	}
	dto := raw.toDTO()

	// CurrentPowerKW should be 0 (default)
	if dto.CurrentPowerKW != 0 {
		t.Errorf("expected CurrentPowerKW 0, got %.2f", dto.CurrentPowerKW)
	}

	// Phases should be nil
	if dto.Phases != nil {
		t.Errorf("expected nil Phases, got %v", dto.Phases)
	}
}

func TestPhaseDetail_Struct(t *testing.T) {
	phase := PhaseDetail{
		Voltage: 230.0,
		Current: 16.0,
		Power:   3680.0,
	}

	if phase.Voltage != 230.0 {
		t.Errorf("expected Voltage 230.0, got %.1f", phase.Voltage)
	}
	if phase.Current != 16.0 {
		t.Errorf("expected Current 16.0, got %.1f", phase.Current)
	}
	if phase.Power != 3680.0 {
		t.Errorf("expected Power 3680.0, got %.1f", phase.Power)
	}
}

func TestWallboxStatus_Struct(t *testing.T) {
	status := WallboxStatus{
		VehicleState:           "Charging",
		ChargingAllowed:        "Yes",
		SetCurrentA:            16,
		CurrentPowerKW:         7.2,
		ChargedSincePlugInKWh:  15.5,
		TotalEnergyLifetimeKWh: 1500.0,
		TemperatureCelsius:     "35.0 °C",
		Phases:                 []PhaseDetail{{Voltage: 230, Current: 16, Power: 3680}},
	}

	if status.VehicleState != "Charging" {
		t.Errorf("expected VehicleState 'Charging', got '%s'", status.VehicleState)
	}
	if status.SetCurrentA != 16 {
		t.Errorf("expected SetCurrentA 16, got %d", status.SetCurrentA)
	}
}
