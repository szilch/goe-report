package cmd

import (
	"testing"
	"time"
)

func TestGetTimeRange_SingleMonth(t *testing.T) {
	startOfPeriod, endOfPeriod, periodLabel, err := getTimeRange("01-2026", "", "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !startOfPeriod.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, startOfPeriod)
	}

	expectedEnd := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)
	if !endOfPeriod.Equal(expectedEnd) {
		t.Errorf("expected end %v, got %v", expectedEnd, endOfPeriod)
	}

	if periodLabel != "01-2026" {
		t.Errorf("expected periodLabel '01-2026', got '%s'", periodLabel)
	}
}

func TestGetTimeRange_MultiMonth(t *testing.T) {
	startOfPeriod, endOfPeriod, periodLabel, err := getTimeRange("", "01-2026", "03-2026")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !startOfPeriod.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, startOfPeriod)
	}

	expectedEnd := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)
	if !endOfPeriod.Equal(expectedEnd) {
		t.Errorf("expected end %v, got %v", expectedEnd, endOfPeriod)
	}

	if periodLabel != "01-2026_to_03-2026" {
		t.Errorf("expected periodLabel '01-2026_to_03-2026', got '%s'", periodLabel)
	}
}

func TestGetTimeRange_InvalidMonthFormat(t *testing.T) {
	_, _, _, err := getTimeRange("2026-01", "", "")

	if err == nil {
		t.Error("expected error for invalid month format, got nil")
	}
}

func TestGetTimeRange_InvalidFromMonthFormat(t *testing.T) {
	_, _, _, err := getTimeRange("", "2026-01", "03-2026")

	if err == nil {
		t.Error("expected error for invalid from-month format, got nil")
	}
}

func TestGetTimeRange_InvalidToMonthFormat(t *testing.T) {
	_, _, _, err := getTimeRange("", "01-2026", "2026-03")

	if err == nil {
		t.Error("expected error for invalid to-month format, got nil")
	}
}

func TestGetTimeRange_ToBeforeFrom(t *testing.T) {
	_, _, _, err := getTimeRange("", "03-2026", "01-2026")

	if err == nil {
		t.Error("expected error when to-month is before from-month, got nil")
	}
}

func TestGetTimeRange_SameFromAndTo(t *testing.T) {
	startOfPeriod, endOfPeriod, periodLabel, err := getTimeRange("", "01-2026", "01-2026")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !startOfPeriod.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, startOfPeriod)
	}

	expectedEnd := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)
	if !endOfPeriod.Equal(expectedEnd) {
		t.Errorf("expected end %v, got %v", expectedEnd, endOfPeriod)
	}

	if periodLabel != "01-2026_to_01-2026" {
		t.Errorf("expected periodLabel '01-2026_to_01-2026', got '%s'", periodLabel)
	}
}

func TestGetTimeRange_December(t *testing.T) {
	startOfPeriod, endOfPeriod, _, err := getTimeRange("12-2025", "", "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedStart := time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC)
	if !startOfPeriod.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, startOfPeriod)
	}

	expectedEnd := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)
	if !endOfPeriod.Equal(expectedEnd) {
		t.Errorf("expected end %v, got %v", expectedEnd, endOfPeriod)
	}
}

func TestGetTimeRange_FullYear(t *testing.T) {
	startOfPeriod, endOfPeriod, periodLabel, err := getTimeRange("", "01-2026", "12-2026")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	if !startOfPeriod.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, startOfPeriod)
	}

	expectedEnd := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)
	if !endOfPeriod.Equal(expectedEnd) {
		t.Errorf("expected end %v, got %v", expectedEnd, endOfPeriod)
	}

	if periodLabel != "01-2026_to_12-2026" {
		t.Errorf("expected periodLabel '01-2026_to_12-2026', got '%s'", periodLabel)
	}
}

func TestGetPreviousMonth(t *testing.T) {
	result := getPreviousMonth()

	// Parse the result
	parsed, err := time.Parse("01-2006", result)
	if err != nil {
		t.Fatalf("getPreviousMonth() returned invalid format: %v", err)
	}

	// Get expected previous month
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	expected := firstOfMonth.AddDate(0, -1, 0)

	if parsed.Year() != expected.Year() || parsed.Month() != expected.Month() {
		t.Errorf("expected %s, got %s", expected.Format("01-2006"), result)
	}
}

func TestGetPreviousMonth_Format(t *testing.T) {
	result := getPreviousMonth()

	// Verify format is MM-YYYY
	if len(result) != 7 {
		t.Errorf("expected length 7, got %d", len(result))
	}

	if result[2] != '-' {
		t.Errorf("expected '-' at position 2, got '%c'", result[2])
	}
}

func TestGetTimeRange_LeapYear(t *testing.T) {
	// 2024 is a leap year
	startOfPeriod, endOfPeriod, _, err := getTimeRange("02-2024", "", "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedStart := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	if !startOfPeriod.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, startOfPeriod)
	}

	// February 2024 has 29 days
	expectedEnd := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)
	if !endOfPeriod.Equal(expectedEnd) {
		t.Errorf("expected end %v, got %v", expectedEnd, endOfPeriod)
	}
}

func TestGetTimeRange_NonLeapYear(t *testing.T) {
	// 2025 is not a leap year
	startOfPeriod, endOfPeriod, _, err := getTimeRange("02-2025", "", "")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedStart := time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)
	if !startOfPeriod.Equal(expectedStart) {
		t.Errorf("expected start %v, got %v", expectedStart, startOfPeriod)
	}

	// February 2025 has 28 days
	expectedEnd := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC).Add(-time.Nanosecond)
	if !endOfPeriod.Equal(expectedEnd) {
		t.Errorf("expected end %v, got %v", expectedEnd, endOfPeriod)
	}
}
