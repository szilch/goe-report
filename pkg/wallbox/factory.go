package wallbox

import (
	"errors"
	"fmt"

	"echarge-report/pkg/config"

	"github.com/spf13/viper"
)

// ErrUnsupportedType is returned when the configured wallbox type is not supported.
var ErrUnsupportedType = errors.New("unsupported wallbox type")

const (
	TypeGoE = "goe"
)

// SupportedTypes returns a list of all supported wallbox type identifiers.
func SupportedTypes() []string {
	return []string{
		TypeGoE,
	}
}

// NewAdapter creates an Adapter by auto-detecting the wallbox type from the
// current configuration. Defaults to TypeGoE if no type is detected.
func NewAdapter() (Adapter, error) {
	wallboxType := DetectWallboxType()

	if wallboxType == "" {
		wallboxType = TypeGoE
	}

	return NewAdapterByType(wallboxType)
}

// DetectWallboxType inspects the current configuration to determine the
// wallbox type. Returns an empty string if no known type is configured.
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

// NewAdapterByType creates an Adapter for the given wallbox type identifier.
// Returns ErrUnsupportedType if the type is not known.
func NewAdapterByType(wallboxType string) (Adapter, error) {
	switch wallboxType {
	case TypeGoE:
		return newGoeAdapter(), nil
	default:
		return nil, fmt.Errorf("%w: %s (supported: %v)", ErrUnsupportedType, wallboxType, SupportedTypes())
	}
}
