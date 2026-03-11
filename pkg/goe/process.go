package goe

import (
	"fmt"
	"strings"
	"goe-report/pkg/formatter"
)

// ChargingLog matches the expected JSON response from the direct_json endpoint
type DirectJsonResp struct {
	Data []ChargingLogRaw `json:"data"`
}

// ChargingLogRaw represents a raw charging log entry as returned by the API
type ChargingLogRaw struct {
	IdChip       interface{} `json:"id_chip"`
	IdChipName   string      `json:"id_chip_name"`
	Start        string      `json:"start"`
	End          string      `json:"end"`
	SecondsTotal string      `json:"seconds_total"`
	Energy       float64     `json:"energy"` // Assumed in kWh
}

// ProcessLogs filters raw charging data by RFID and maps it into the formatter.SessionData struct.
func ProcessLogs(data *DirectJsonResp, chipIdsFlag string, kwhPrice float64) (sessions []formatter.SessionData, totalEnergy, totalPrice float64, totalSessions int) {
	for _, session := range data.Data {
		var idChipStr string
		if session.IdChip != nil {
			idChipStr = fmt.Sprintf("%v", session.IdChip)
		}

		matched := false
		if chipIdsFlag == "" {
			matched = true
		} else {
			validIds := strings.Split(chipIdsFlag, ",")
			for _, vid := range validIds {
				v := strings.TrimSpace(vid)
				if idChipStr == v || session.IdChipName == v {
					matched = true
					break
				}
			}
		}

		if !matched {
			continue
		}

		sessionPrice := session.Energy * kwhPrice

		totalEnergy += session.Energy
		totalSessions++
		totalPrice += sessionPrice

		sessions = append(sessions, formatter.SessionData{
			StartDate: session.Start,
			EndDate:   session.End,
			Duration:  session.SecondsTotal,
			Energy:    session.Energy,
			Price:     sessionPrice,
			RFID:      idChipStr,
		})
	}

	return sessions, totalEnergy, totalPrice, totalSessions
}
