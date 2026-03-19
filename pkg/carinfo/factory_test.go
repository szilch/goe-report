package carinfo

import (
	"echarge-report/pkg/config"
	"testing"

	"github.com/spf13/viper"
)

func TestSupportedTypes(t *testing.T) {
	types := SupportedTypes()

	if len(types) == 0 {
		t.Error("SupportedTypes() should return at least one type")
	}

	found := false
	for _, pType := range types {
		if pType == TypeHomeAssistant {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("SupportedTypes() should contain TypeHomeAssistant (%s), got: %v", TypeHomeAssistant, types)
	}
}

func TestNewProviderByType_HomeAssistant(t *testing.T) {
	cfg := Config{
		ProviderType: TypeHomeAssistant,
		HAWsHost:     "ws://ha.local",
		HAToken:      "test-token",
	}

	provider, err := NewProviderByType(TypeHomeAssistant, cfg)

	if err != nil {
		t.Errorf("NewProviderByType(TypeHomeAssistant) returned error: %v", err)
	}
	if provider == nil {
		t.Error("NewProviderByType(TypeHomeAssistant) returned nil provider")
	}
	if provider != nil && provider.GetType() != TypeHomeAssistant {
		t.Errorf("Provider type should be %s, got: %s", TypeHomeAssistant, provider.GetType())
	}
}

func TestNewProvider_NoConfig(t *testing.T) {
	provider, err := NewProvider(Config{})

	if err != nil {
		t.Errorf("NewProvider() should not fail with no config, got error: %v", err)
	}
	if provider != nil {
		t.Error("NewProvider() should return nil provider when no config is set")
	}
}

func TestDetectProviderType_EnvVars(t *testing.T) {
	defer viper.Reset()

	// Simulate environment variable being set without the branch key
	viper.Set(config.KeyHAWsHost, "ws://ha.local")

	providerType := DetectProviderType()

	if providerType != TypeHomeAssistant {
		t.Errorf("Expected %s detected via leaf key, got: %s", TypeHomeAssistant, providerType)
	}
}

func TestDetectProviderType_BranchKey(t *testing.T) {
	defer viper.Reset()

	// Simulate config file branch key being set
	viper.Set("smarthome.homeassistant", map[string]string{"foo": "bar"})

	providerType := DetectProviderType()

	if providerType != TypeHomeAssistant {
		t.Errorf("Expected %s detected via branch key, got: %s", TypeHomeAssistant, providerType)
	}
}
