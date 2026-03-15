package carinfo

// Provider defines the interface for fetching car-related information.
type Provider interface {
	// GetMileage returns the current mileage of the car.
	GetMileage() (string, error)

	// GetType returns the provider type.
	GetType() string
}
