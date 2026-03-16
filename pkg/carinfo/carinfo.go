package carinfo

import "time"

// Provider defines the interface for fetching car-related information.
type Provider interface {
	// GetMileage returns the current mileage of the car.
	GetMileage() (string, error)

	// GetMileageAt returns the mileage of the car at a specific point in time.
	GetMileageAt(t time.Time) (string, error)

	// GetType returns the provider type.
	GetType() string
}
