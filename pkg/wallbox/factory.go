package wallbox

import (
	"fmt"
	"goe-report/pkg/config"
	goeAdapter "goe-report/pkg/wallbox/goe"
	"goe-report/pkg/wallbox/types"

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
// If no type is configured, it defaults to "goe" for backward compatibility.
func NewAdapter() (types.Adapter, error) {
	wallboxType := viper.GetString(config.KeyWallboxType)

	// Default to "goe" for backward compatibility
	if wallboxType == "" {
		wallboxType = TypeGoE
	}

	return NewAdapterByType(wallboxType)
}

// NewAdapterByType creates a new wallbox adapter for the specified type.
// Returns an error if the wallbox type is not supported.
func NewAdapterByType(wallboxType string) (types.Adapter, error) {
	switch wallboxType {
	case TypeGoE:
		return goeAdapter.NewAdapter(), nil
	// Add new adapter cases here, e.g.:
	// case TypeEasee:
	//     return easeeAdapter.NewAdapter(), nil
	default:
		return nil, fmt.Errorf("unsupported wallbox type: %s. Supported types: %v", wallboxType, SupportedTypes())
	}
}
