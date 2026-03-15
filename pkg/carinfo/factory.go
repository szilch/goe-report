package carinfo

import (
	"echarge-report/pkg/config"
	"fmt"

	"github.com/spf13/viper"
)

const (
	TypeHomeAssistant = "homeassistant"
)

func SupportedTypes() []string {
	return []string{
		TypeHomeAssistant,
	}
}

func NewProvider() (Provider, error) {
	providerType := DetectProviderType()

	if providerType == "" {
		// Default or no provider configured
		return nil, nil
	}

	return NewProviderByType(providerType)
}

func DetectProviderType() string {
	for _, t := range SupportedTypes() {
		if viper.IsSet(fmt.Sprintf("%s.%s", config.KeySmarthome, t)) {
			return t
		}
	}
	return ""
}

func NewProviderByType(providerType string) (Provider, error) {
	switch providerType {
	case TypeHomeAssistant:
		return NewHomeAssistantProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported smarthome provider: %s. Supported types: %v", providerType, SupportedTypes())
	}
}
