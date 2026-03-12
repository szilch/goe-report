package wallbox

import "echarge-report/pkg/wallbox/types"

// Re-export types for backward compatibility and convenience
type (
	PhaseDetail      = types.PhaseDetail
	Status           = types.Status
	ChargingSession  = types.ChargingSession
	ChargingResponse = types.ChargingResponse
	Adapter          = types.Adapter
)
