package carinfo

import (
	"errors"
	"fmt"

	"echarge-report/pkg/config"

	"github.com/spf13/viper"
)

// ErrUnsupportedProvider is returned when the configured smarthome provider is not supported.
var ErrUnsupportedProvider = errors.New("unsupported smarthome provider")

const (
	TypeHomeAssistant = "homeassistant"
)

// SupportedTypes returns a list of all supported smarthome provider types.
func SupportedTypes() []string {
	return []string{
		TypeHomeAssistant,
	}
}

// NewProvider creates a Provider based on the given configuration.
// If no type is specified, it returns nil without an error.
func NewProvider(cfg Config) (Provider, error) {
	if cfg.ProviderType == "" {
		// Default or no provider configured
		return nil, nil
	}

	return NewProviderByType(cfg.ProviderType, cfg)
}

// DetectProviderType inspects the current configuration via viper to determine
// the configured smarthome provider type.
func DetectProviderType() string {
	for _, t := range SupportedTypes() {
		// Check for the branch key (works for file config)
		if viper.IsSet(fmt.Sprintf("%s.%s", config.KeySmarthome, t)) {
			return t
		}
		
		// Fallback: check for provider specific leaf keys (needed for environment variables)
		switch t {
		case TypeHomeAssistant:
			if viper.IsSet(config.KeyHAWsHost) || viper.IsSet(config.KeyHAToken) {
				return TypeHomeAssistant
			}
		}
	}
	return ""
}

// NewProviderByType creates a specific Provider instance based on the providerType.
// Returns ErrUnsupportedProvider if the type is unknown.
func NewProviderByType(providerType string, cfg Config) (Provider, error) {
	switch providerType {
	case TypeHomeAssistant:
		return NewHomeAssistantProvider(cfg), nil
	default:
		return nil, fmt.Errorf("%w: %s (supported: %v)", ErrUnsupportedProvider, providerType, SupportedTypes())
	}
}
