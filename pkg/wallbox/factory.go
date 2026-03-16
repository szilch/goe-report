package wallbox

import (
	"echarge-report/pkg/config"
	"fmt"

	"github.com/spf13/viper"
)

const (
	TypeGoE = "goe"
)

func SupportedTypes() []string {
	return []string{
		TypeGoE,
	}
}

func NewAdapter() (Adapter, error) {
	wallboxType := DetectWallboxType()

	if wallboxType == "" {
		wallboxType = TypeGoE
	}

	return NewAdapterByType(wallboxType)
}

func DetectWallboxType() string {
	for _, t := range SupportedTypes() {
		// Check for the branch key (works for file config)
		if viper.IsSet(fmt.Sprintf("%s.%s", config.KeyWallbox, t)) {
			return t
		}

		// Fallback: check for adapter specific leaf keys (needed for environment variables)
		switch t {
		case TypeGoE:
			if viper.IsSet(config.KeyWallboxGoeCloudSerial) || viper.IsSet(config.KeyWallboxGoeLocalApiUrl) {
				return TypeGoE
			}
		}
	}
	return ""
}

func NewAdapterByType(wallboxType string) (Adapter, error) {
	switch wallboxType {
	case TypeGoE:
		return newGoeAdapter(), nil
	default:
		return nil, fmt.Errorf("unsupported wallbox type: %s. Supported types: %v", wallboxType, SupportedTypes())
	}
}
