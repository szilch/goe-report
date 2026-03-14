package wallbox

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

	// Check that TypeGoE is in the list
	found := false
	for _, wType := range types {
		if wType == TypeGoE {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("SupportedTypes() should contain TypeGoE (%s), got: %v", TypeGoE, types)
	}
}

func TestTypeConstants(t *testing.T) {
	// Verify type constants are not empty
	if TypeGoE == "" {
		t.Error("TypeGoE constant should not be empty")
	}

	// Verify TypeGoE has expected value
	if TypeGoE != "goe" {
		t.Errorf("TypeGoE should be 'goe', got: %s", TypeGoE)
	}
}

func TestNewAdapterByType_GoE(t *testing.T) {
	// Setup minimal config required for go-e adapter
	viper.Set(config.KeyWallboxGoeCloudSerial, "test-serial")
	viper.Set(config.KeyWallboxGoeCloudToken, "test-token")
	defer viper.Reset()

	adapter, err := NewAdapterByType(TypeGoE)

	if err != nil {
		t.Errorf("NewAdapterByType(TypeGoE) returned error: %v", err)
	}
	if adapter == nil {
		t.Error("NewAdapterByType(TypeGoE) returned nil adapter")
	}
	if adapter != nil && adapter.GetType() != TypeGoE {
		t.Errorf("Adapter type should be %s, got: %s", TypeGoE, adapter.GetType())
	}
}

func TestNewAdapterByType_Unsupported(t *testing.T) {
	adapter, err := NewAdapterByType("unsupported-type")

	if err == nil {
		t.Error("NewAdapterByType('unsupported-type') should return an error")
	}
	if adapter != nil {
		t.Error("NewAdapterByType('unsupported-type') should return nil adapter")
	}
}

func TestNewAdapter_Default(t *testing.T) {
	// Setup goe configuration
	viper.Set(config.KeyWallboxGoeCloudSerial, "test-serial")
	viper.Set(config.KeyWallboxGoeCloudToken, "test-token")
	defer viper.Reset()

	adapter, err := NewAdapter()

	if err != nil {
		t.Errorf("NewAdapter() failed: %v", err)
	}
	if adapter == nil {
		t.Error("NewAdapter() returned nil adapter")
	}
	if adapter != nil && adapter.GetType() != TypeGoE {
		t.Errorf("Detected adapter type should be %s, got: %s", TypeGoE, adapter.GetType())
	}
}

func TestNewAdapter_NoConfig(t *testing.T) {
	// No configuration set
	defer viper.Reset()

	adapter, err := NewAdapter()

	if err != nil {
		t.Errorf("NewAdapter() should not fail even with no config (defaults to goe), got error: %v", err)
	}
	if adapter == nil {
		t.Error("NewAdapter() returned nil adapter")
	}
	if adapter != nil && adapter.GetType() != TypeGoE {
		t.Errorf("Default adapter type should be %s, got: %s", TypeGoE, adapter.GetType())
	}
}

func TestSupportedTypes_ContainsAllConstants(t *testing.T) {
	types := SupportedTypes()

	// Build a map for quick lookup
	typeMap := make(map[string]bool)
	for _, wType := range types {
		typeMap[wType] = true
	}

	// Verify all known type constants are in SupportedTypes
	knownTypes := []string{TypeGoE}
	for _, knownType := range knownTypes {
		if !typeMap[knownType] {
			t.Errorf("SupportedTypes() should contain %s", knownType)
		}
	}
}
