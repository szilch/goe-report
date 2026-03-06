package goe

import "fmt"

// PhaseDetail represents the electrical details of a single charging phase.
type PhaseDetail struct {
	Voltage float64
	Current float64
	Power   float64
}

// WallboxStatus is the refined Data Transfer Object providing human-readable values.
type WallboxStatus struct {
	VehicleState           string
	ChargingAllowed        string
	SetCurrentA            int
	CurrentPowerKW         float64
	ChargedSincePlugInKWh  float64
	TotalEnergyLifetimeKWh float64
	TemperatureCelsius     string
	Phases                 []PhaseDetail
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

// toDTO converts the raw JSON API response into the clean WallboxStatus DTO.
func (raw *rawStatusData) toDTO() *WallboxStatus {
	dto := &WallboxStatus{
		SetCurrentA:            raw.Amp,
		ChargedSincePlugInKWh:  raw.Wh / 1000.0,
		TotalEnergyLifetimeKWh: raw.Eto / 1000.0,
	}

	// Interpret the 'car' state
	switch raw.Car {
	case 1:
		dto.VehicleState = "Idle (not connected)"
	case 2:
		dto.VehicleState = "Charging"
	case 3:
		dto.VehicleState = "Waiting for car"
	case 4:
		dto.VehicleState = "Charging complete"
	case 5:
		dto.VehicleState = "Error"
	default:
		dto.VehicleState = "Unknown"
	}

	// Allowed state
	dto.ChargingAllowed = "No"
	if raw.Alw {
		dto.ChargingAllowed = "Yes"
	}

	// Calculate total power
	numNrg := len(raw.Nrg)
	if numNrg >= 12 {
		dto.CurrentPowerKW = raw.Nrg[11] / 1000.0
	} else if numNrg >= 4 {
		// Fallback for smaller arrays (not implemented here since index 11 is standard for v3/v4)
	}

	// Temperature
	dto.TemperatureCelsius = "N/A"
	if len(raw.Tma) > 0 {
		dto.TemperatureCelsius = fmt.Sprintf("%.1f °C", raw.Tma[0])
	}

	// Phase details
	if numNrg >= 10 {
		dto.Phases = []PhaseDetail{
			{Voltage: raw.Nrg[0], Current: raw.Nrg[4], Power: raw.Nrg[7]},
			{Voltage: raw.Nrg[1], Current: raw.Nrg[5], Power: raw.Nrg[8]},
			{Voltage: raw.Nrg[2], Current: raw.Nrg[6], Power: raw.Nrg[9]},
		}
	}

	return dto
}
