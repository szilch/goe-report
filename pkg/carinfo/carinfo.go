package carinfo

import "time"

// Provider defines the interface for fetching car information like mileage.
type Provider interface {
	GetMileage() (int, error)
	GetMileageAt(t time.Time) (int, error)
	GetType() string
}

// Config holds the configuration required for creating a carinfo Provider.
type Config struct {
	ProviderType    string
	HAWsHost        string
	HAToken         string
	HAMileageSensor string
}
