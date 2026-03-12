package types

// PhaseDetail represents the electrical details of a single charging phase.
type PhaseDetail struct {
	Voltage float64
	Current float64
	Power   float64
}

// Status is the generalized wallbox status DTO providing human-readable values.
// Specific adapters map their proprietary data to this common structure.
type Status struct {
	VehicleState           string
	ChargingAllowed        string
	SetCurrentA            int
	CurrentPowerKW         float64
	ChargedSincePlugInKWh  float64
	TotalEnergyLifetimeKWh float64
	TemperatureCelsius     string
	Phases                 []PhaseDetail
}

// ChargingSession represents a single charging session in a generalized format.
type ChargingSession struct {
	IdChip       interface{}
	IdChipName   string
	Start        string
	End          string
	SecondsTotal string
	Energy       float64 // in kWh
}

// ChargingResponse is the generalized response containing multiple charging sessions.
type ChargingResponse struct {
	Data []ChargingSession
}

// Adapter defines the common interface that all wallbox implementations must satisfy.
// This allows the application to work with any wallbox type in a unified way.
type Adapter interface {
	// FetchChargingData fetches charging session data for the given time range.
	// fromMs and toMs are Unix timestamps in milliseconds.
	FetchChargingData(fromMs, toMs int64) (*ChargingResponse, error)

	// GetStatus fetches the current status of the wallbox.
	GetStatus() (*Status, error)

	// GetType returns the type identifier of the wallbox (e.g., "goe", "easee", etc.)
	GetType() string
}
