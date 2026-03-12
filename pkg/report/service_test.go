package report

import (
	"echarge-report/pkg/config"
	"echarge-report/pkg/wallbox"
	"math"
	"testing"
	"time"

	"github.com/spf13/viper"
)

// MockWallboxAdapter provides a mock implementation of WallboxAdapter.
type MockWallboxAdapter struct {
	FetchChargingDataFunc func(fromMs, toMs int64) (*wallbox.ChargingResponse, error)
}

func (m *MockWallboxAdapter) FetchChargingData(fromMs, toMs int64) (*wallbox.ChargingResponse, error) {
	if m.FetchChargingDataFunc != nil {
		return m.FetchChargingDataFunc(fromMs, toMs)
	}
	return &wallbox.ChargingResponse{Data: []wallbox.ChargingSession{}}, nil
}

// MockHAService provides a mock implementation of HAService.
type MockHAService struct {
	GetSensorValueFunc func(sensorID string) (string, error)
}

func (m *MockHAService) GetSensorValue(sensorID string) (string, error) {
	if m.GetSensorValueFunc != nil {
		return m.GetSensorValueFunc(sensorID)
	}
	return "mock-mileage", nil
}

func TestService_GenerateReportData(t *testing.T) {
	// Setup standard viper config for test
	viper.Set(config.KeyWallboxSerial, "123456")
	viper.Set(config.KeyLicensePlate, "TEST-123")
	viper.Set(config.KeyKwhPrice, 0.35)
	viper.Set(config.KeyWallboxChipIds, "")

	// Clean up config after max
	defer viper.Reset()

	mockAdapter := &MockWallboxAdapter{
		FetchChargingDataFunc: func(fromMs, toMs int64) (*wallbox.ChargingResponse, error) {
			return &wallbox.ChargingResponse{
				Data: []wallbox.ChargingSession{
					{
						IdChip:       "chip-1",
						IdChipName:   "TestChip",
						Start:        "01.01.2026 10:00:00",
						End:          "01.01.2026 12:00:00",
						SecondsTotal: "7200",
						Energy:       20.0,
					},
				},
			}, nil
		},
	}

	mockHA := &MockHAService{
		GetSensorValueFunc: func(sensorID string) (string, error) {
			return "50000", nil
		},
	}
	viper.Set(config.KeyHAMilageSensor, "sensor.test_mileage")

	s := NewService(mockAdapter, mockHA)

	report, err := s.GenerateReportData("01-2026", "", "")
	if err != nil {
		t.Fatalf("GenerateReportData failed: %v", err)
	}

	if report.SerialNumber != "123456" {
		t.Errorf("Expected SerialNumber 123456, got %s", report.SerialNumber)
	}
	if report.LicensePlate != "TEST-123" {
		t.Errorf("Expected LicensePlate TEST-123, got %s", report.LicensePlate)
	}
	if report.KwhPrice != 0.35 {
		t.Errorf("Expected KwhPrice 0.35, got %f", report.KwhPrice)
	}
	if report.TotalSessions != 1 {
		t.Errorf("Expected TotalSessions 1, got %d", report.TotalSessions)
	}
	if report.TotalEnergy != 20.0 {
		t.Errorf("Expected TotalEnergy 20.0, got %f", report.TotalEnergy)
	}
	if math.Abs(report.TotalPrice-7.0) > 0.001 { // 20.0 * 0.35
		t.Errorf("Expected TotalPrice 7.0, got %f", report.TotalPrice)
	}
	if report.Mileage != "50000" {
		t.Errorf("Expected Mileage 50000, got %s", report.Mileage)
	}
	if report.PeriodLabel != "01-2026" {
		t.Errorf("Expected PeriodLabel 01-2026, got %s", report.PeriodLabel)
	}
}

func TestService_getPreviousMonth(t *testing.T) {
	s := NewService(nil, nil)
	prevMonth := s.getPreviousMonth()

	// Parse it back to verify format
	parsed, err := time.Parse("01-2006", prevMonth)
	if err != nil {
		t.Fatalf("getPreviousMonth returned invalid format: %v", err)
	}

	// Verify it's actually the previous month
	now := time.Now()
	expectedMonth := now.Month() - 1
	expectedYear := now.Year()
	if expectedMonth == 0 {
		expectedMonth = 12
		expectedYear--
	}

	if parsed.Month() != expectedMonth || parsed.Year() != expectedYear {
		t.Errorf("Expected %02d-%d, got %02d-%d", expectedMonth, expectedYear, parsed.Month(), parsed.Year())
	}
}

func TestService_getTimeRange(t *testing.T) {
	s := NewService(nil, nil)

	tests := []struct {
		name          string
		monthFlag     string
		fromMonthFlag string
		toMonthFlag   string
		expectError   bool
		expectedLabel string
	}{
		{
			name:          "single month",
			monthFlag:     "01-2026",
			expectError:   false,
			expectedLabel: "01-2026",
		},
		{
			name:        "single month invalid",
			monthFlag:   "invalid",
			expectError: true,
		},
		{
			name:          "range",
			fromMonthFlag: "01-2026",
			toMonthFlag:   "03-2026",
			expectError:   false,
			expectedLabel: "01-2026_to_03-2026",
		},
		{
			name:          "range invalid from",
			fromMonthFlag: "invalid",
			toMonthFlag:   "03-2026",
			expectError:   true,
		},
		{
			name:          "range invalid to",
			fromMonthFlag: "01-2026",
			toMonthFlag:   "invalid",
			expectError:   true,
		},
		{
			name:          "range inverted",
			fromMonthFlag: "03-2026",
			toMonthFlag:   "01-2026",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start, end, label, err := s.getTimeRange(tt.monthFlag, tt.fromMonthFlag, tt.toMonthFlag)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if label != tt.expectedLabel {
				t.Errorf("Expected label %s, got %s", tt.expectedLabel, label)
			}

			if start.IsZero() || end.IsZero() {
				t.Errorf("Expected non-zero start and end times")
			}

			if end.Before(start) {
				t.Errorf("End time is before start time")
			}
		})
	}
}

func TestService_processLogs(t *testing.T) {
	s := NewService(nil, nil)
	kwhPrice := 0.30

	data := &wallbox.ChargingResponse{
		Data: []wallbox.ChargingSession{
			{
				IdChip:       "12345",
				IdChipName:   "Chip1",
				Start:        "01.01.2026 10:00:00",
				End:          "01.01.2026 12:00:00",
				SecondsTotal: "7200",
				Energy:       10.0,
			},
			{
				IdChip:       67890, // Testing with float/int type mapping
				IdChipName:   "Chip2",
				Start:        "02.01.2026 10:00:00",
				End:          "02.01.2026 12:00:00",
				SecondsTotal: "7200",
				Energy:       20.0,
			},
			{
				IdChip:       nil,
				IdChipName:   "",
				Start:        "03.01.2026 10:00:00",
				End:          "03.01.2026 12:00:00",
				SecondsTotal: "7200",
				Energy:       30.0,
			},
		},
	}

	tests := []struct {
		name             string
		chipIdsFlag      string
		expectedSessions int
		expectedEnergy   float64
		expectedPrice    float64
	}{
		{
			name:             "no filter",
			chipIdsFlag:      "",
			expectedSessions: 3,
			expectedEnergy:   60.0,
			expectedPrice:    60.0 * kwhPrice,
		},
		{
			name:             "single chip by id string",
			chipIdsFlag:      "12345",
			expectedSessions: 1,
			expectedEnergy:   10.0,
			expectedPrice:    10.0 * kwhPrice,
		},
		{
			name:             "single chip by id int",
			chipIdsFlag:      "67890",
			expectedSessions: 1,
			expectedEnergy:   20.0,
			expectedPrice:    20.0 * kwhPrice,
		},
		{
			name:             "single chip by name",
			chipIdsFlag:      "Chip1",
			expectedSessions: 1,
			expectedEnergy:   10.0,
			expectedPrice:    10.0 * kwhPrice,
		},
		{
			name:             "multiple chips",
			chipIdsFlag:      "12345, 67890", // With space to test TrimSpace
			expectedSessions: 2,
			expectedEnergy:   30.0,
			expectedPrice:    30.0 * kwhPrice,
		},
		{
			name:             "unknown chip",
			chipIdsFlag:      "99999",
			expectedSessions: 0,
			expectedEnergy:   0.0,
			expectedPrice:    0.0 * kwhPrice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessions, totalEnergy, totalPrice, totalSessions := s.processLogs(data, tt.chipIdsFlag, kwhPrice)

			if len(sessions) != tt.expectedSessions {
				t.Errorf("Expected %d sessions, got %d", tt.expectedSessions, len(sessions))
			}
			if totalSessions != tt.expectedSessions {
				t.Errorf("Expected totalSessions %d, got %d", tt.expectedSessions, totalSessions)
			}
			if math.Abs(totalEnergy-tt.expectedEnergy) > 0.001 {
				t.Errorf("Expected totalEnergy %.2f, got %.2f", tt.expectedEnergy, totalEnergy)
			}
			if math.Abs(totalPrice-tt.expectedPrice) > 0.001 {
				t.Errorf("Expected totalPrice %.2f, got %.2f", tt.expectedPrice, totalPrice)
			}
		})
	}
}
