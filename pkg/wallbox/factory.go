package wallbox

import (
	"echarge-report/pkg/config"
	"fmt"

	"github.com/spf13/viper"
)

// WallboxType constants for supported wallbox types.
const (
	TypeGoE = "goe"
	// Add new wallbox types here, e.g.:
	// TypeEasee = "easee"
	// TypeKeba = "keba"
)

// SupportedTypes returns a list of all supported wallbox types.
func SupportedTypes() []string {
	return []string{
		TypeGoE,
		// Add new types here
	}
}

// NewAdapter creates a new wallbox adapter based on the configured wallbox type.
// If no type is explicitly configured, it attempts to detect the type from the configuration structure.
func NewAdapter() (Adapter, error) {
	wallboxType := viper.GetString(config.KeyWallboxType)

	if wallboxType == "" {
		wallboxType = DetectWallboxType()
	}

	// Default to "goe" for backward compatibility if still not found
	if wallboxType == "" {
		wallboxType = TypeGoE
	}

	return NewAdapterByType(wallboxType)
}

// DetectWallboxType attempts to determine the wallbox type based on the presence of keys in the "wallbox" section.
func DetectWallboxType() string {
	for _, t := range SupportedTypes() {
		// Check if the wallbox.<type> key exists in configuration
		if viper.IsSet(fmt.Sprintf("%s.%s", config.KeyWallbox, t)) {
			return t
		}
	}
	return ""
}

// NewAdapterByType creates a new wallbox adapter for the specified type.
// Returns an error if the wallbox type is not supported.
func NewAdapterByType(wallboxType string) (Adapter, error) {
	switch wallboxType {
	case TypeGoE:
		return newGoeAdapter(), nil
	// Add new adapter cases here, e.g.:
	// case TypeEasee:
	//     return newEaseeAdapter(), nil
	default:
		return nil, fmt.Errorf("unsupported wallbox type: %s. Supported types: %v", wallboxType, SupportedTypes())
	}
}

