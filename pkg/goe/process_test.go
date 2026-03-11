package goe

import (
	"testing"

	"goe-report/pkg/formatter"
)

func TestProcessLogs_NoFilter(t *testing.T) {
	data := &DirectJsonResp{
		Data: []ChargingLogRaw{
			{
				IdChip:       "12345",
				IdChipName:   "Card1",
				Start:        "2026-01-15 10:00:00",
				End:          "2026-01-15 12:00:00",
				SecondsTotal: "7200",
				Energy:       10.5,
			},
			{
				IdChip:       "67890",
				IdChipName:   "Card2",
				Start:        "2026-01-16 14:00:00",
				End:          "2026-01-16 15:30:00",
				SecondsTotal: "5400",
				Energy:       7.2,
			},
		},
	}

	sessions, totalEnergy, totalPrice, totalSessions := ProcessLogs(data, "", 0.35)

	if totalSessions != 2 {
		t.Errorf("expected 2 sessions, got %d", totalSessions)
	}

	expectedEnergy := 17.7
	if totalEnergy != expectedEnergy {
		t.Errorf("expected total energy %.2f, got %.2f", expectedEnergy, totalEnergy)
	}

	expectedPrice := expectedEnergy * 0.35
	if totalPrice != expectedPrice {
		t.Errorf("expected total price %.2f, got %.2f", expectedPrice, totalPrice)
	}

	if len(sessions) != 2 {
		t.Errorf("expected 2 session entries, got %d", len(sessions))
	}
}

func TestProcessLogs_FilterByChipId(t *testing.T) {
	data := &DirectJsonResp{
		Data: []ChargingLogRaw{
			{
				IdChip:       "12345",
				IdChipName:   "Card1",
				Start:        "2026-01-15 10:00:00",
				End:          "2026-01-15 12:00:00",
				SecondsTotal: "7200",
				Energy:       10.5,
			},
			{
				IdChip:       "67890",
				IdChipName:   "Card2",
				Start:        "2026-01-16 14:00:00",
				End:          "2026-01-16 15:30:00",
				SecondsTotal: "5400",
				Energy:       7.2,
			},
		},
	}

	sessions, totalEnergy, totalPrice, totalSessions := ProcessLogs(data, "12345", 0.40)

	if totalSessions != 1 {
		t.Errorf("expected 1 session, got %d", totalSessions)
	}

	if totalEnergy != 10.5 {
		t.Errorf("expected total energy 10.5, got %.2f", totalEnergy)
	}

	expectedPrice := 10.5 * 0.40
	if totalPrice != expectedPrice {
		t.Errorf("expected total price %.2f, got %.2f", expectedPrice, totalPrice)
	}

	if len(sessions) != 1 {
		t.Errorf("expected 1 session entry, got %d", len(sessions))
	}

	if sessions[0].RFID != "12345" {
		t.Errorf("expected RFID '12345', got '%s'", sessions[0].RFID)
	}
}

func TestProcessLogs_FilterByChipName(t *testing.T) {
	data := &DirectJsonResp{
		Data: []ChargingLogRaw{
			{
				IdChip:       "12345",
				IdChipName:   "Card1",
				Start:        "2026-01-15 10:00:00",
				End:          "2026-01-15 12:00:00",
				SecondsTotal: "7200",
				Energy:       10.5,
			},
			{
				IdChip:       "67890",
				IdChipName:   "Card2",
				Start:        "2026-01-16 14:00:00",
				End:          "2026-01-16 15:30:00",
				SecondsTotal: "5400",
				Energy:       7.2,
			},
		},
	}

	sessions, _, _, totalSessions := ProcessLogs(data, "Card2", 0.35)

	if totalSessions != 1 {
		t.Errorf("expected 1 session, got %d", totalSessions)
	}

	if len(sessions) != 1 {
		t.Errorf("expected 1 session entry, got %d", len(sessions))
	}

	if sessions[0].RFID != "67890" {
		t.Errorf("expected RFID '67890', got '%s'", sessions[0].RFID)
	}
}

func TestProcessLogs_MultipleChipIdsFilter(t *testing.T) {
	data := &DirectJsonResp{
		Data: []ChargingLogRaw{
			{
				IdChip:       "12345",
				IdChipName:   "Card1",
				Start:        "2026-01-15 10:00:00",
				End:          "2026-01-15 12:00:00",
				SecondsTotal: "7200",
				Energy:       10.0,
			},
			{
				IdChip:       "67890",
				IdChipName:   "Card2",
				Start:        "2026-01-16 14:00:00",
				End:          "2026-01-16 15:30:00",
				SecondsTotal: "5400",
				Energy:       5.0,
			},
			{
				IdChip:       "99999",
				IdChipName:   "Card3",
				Start:        "2026-01-17 08:00:00",
				End:          "2026-01-17 09:00:00",
				SecondsTotal: "3600",
				Energy:       3.0,
			},
		},
	}

	sessions, totalEnergy, _, totalSessions := ProcessLogs(data, "12345,67890", 0.35)

	if totalSessions != 2 {
		t.Errorf("expected 2 sessions, got %d", totalSessions)
	}

	if totalEnergy != 15.0 {
		t.Errorf("expected total energy 15.0, got %.2f", totalEnergy)
	}

	if len(sessions) != 2 {
		t.Errorf("expected 2 session entries, got %d", len(sessions))
	}
}

func TestProcessLogs_EmptyData(t *testing.T) {
	data := &DirectJsonResp{
		Data: []ChargingLogRaw{},
	}

	sessions, totalEnergy, totalPrice, totalSessions := ProcessLogs(data, "", 0.35)

	if totalSessions != 0 {
		t.Errorf("expected 0 sessions, got %d", totalSessions)
	}

	if totalEnergy != 0 {
		t.Errorf("expected total energy 0, got %.2f", totalEnergy)
	}

	if totalPrice != 0 {
		t.Errorf("expected total price 0, got %.2f", totalPrice)
	}

	if len(sessions) != 0 {
		t.Errorf("expected 0 session entries, got %d", len(sessions))
	}
}

func TestProcessLogs_NoMatchingChipId(t *testing.T) {
	data := &DirectJsonResp{
		Data: []ChargingLogRaw{
			{
				IdChip:       "12345",
				IdChipName:   "Card1",
				Start:        "2026-01-15 10:00:00",
				End:          "2026-01-15 12:00:00",
				SecondsTotal: "7200",
				Energy:       10.5,
			},
		},
	}

	sessions, _, _, totalSessions := ProcessLogs(data, "99999", 0.35)

	if totalSessions != 0 {
		t.Errorf("expected 0 sessions, got %d", totalSessions)
	}

	if len(sessions) != 0 {
		t.Errorf("expected 0 session entries, got %d", len(sessions))
	}
}

func TestProcessLogs_NilChipId(t *testing.T) {
	data := &DirectJsonResp{
		Data: []ChargingLogRaw{
			{
				IdChip:       nil, // nil chip ID
				IdChipName:   "Anonymous",
				Start:        "2026-01-15 10:00:00",
				End:          "2026-01-15 12:00:00",
				SecondsTotal: "7200",
				Energy:       10.5,
			},
		},
	}

	// Without filter, should match
	sessions, _, _, totalSessions := ProcessLogs(data, "", 0.35)

	if totalSessions != 1 {
		t.Errorf("expected 1 session, got %d", totalSessions)
	}

	if len(sessions) != 1 {
		t.Errorf("expected 1 session entry, got %d", len(sessions))
	}

	if sessions[0].RFID != "" {
		t.Errorf("expected empty RFID, got '%s'", sessions[0].RFID)
	}
}

func TestProcessLogs_SessionDataMapping(t *testing.T) {
	data := &DirectJsonResp{
		Data: []ChargingLogRaw{
			{
				IdChip:       "12345",
				IdChipName:   "Card1",
				Start:        "2026-01-15 10:00:00",
				End:          "2026-01-15 12:00:00",
				SecondsTotal: "7200",
				Energy:       10.5,
			},
		},
	}

	sessions, _, _, _ := ProcessLogs(data, "", 0.40)

	expected := formatter.SessionData{
		StartDate: "2026-01-15 10:00:00",
		EndDate:   "2026-01-15 12:00:00",
		Duration:  "7200",
		Energy:    10.5,
		Price:     10.5 * 0.40,
		RFID:      "12345",
	}

	if sessions[0] != expected {
		t.Errorf("expected session data %+v, got %+v", expected, sessions[0])
	}
}

func TestProcessLogs_WhitespaceInFilter(t *testing.T) {
	data := &DirectJsonResp{
		Data: []ChargingLogRaw{
			{
				IdChip:       "12345",
				IdChipName:   "Card1",
				Start:        "2026-01-15 10:00:00",
				End:          "2026-01-15 12:00:00",
				SecondsTotal: "7200",
				Energy:       10.5,
			},
		},
	}

	// Filter with spaces
	sessions, _, _, totalSessions := ProcessLogs(data, " 12345 , 67890 ", 0.35)

	if totalSessions != 1 {
		t.Errorf("expected 1 session, got %d", totalSessions)
	}

	if len(sessions) != 1 {
		t.Errorf("expected 1 session entry, got %d", len(sessions))
	}
}
