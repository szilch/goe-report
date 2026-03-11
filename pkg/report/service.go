package report

import (
	"fmt"
	"goe-report/pkg/config"
	"goe-report/pkg/goe"
	"goe-report/pkg/models"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// GoeClient defines the interface for communicating with the go-e API.
type GoeClient interface {
	FetchChargingData(fromMs, toMs int64) (*goe.DirectJsonResp, error)
}

// HAService defines the interface for communicating with Home Assistant.
type HAService interface {
	GetSensorValue(sensorID string) (string, error)
}

// Service provides functionality for generating charging reports.
type Service struct {
	goeClient GoeClient
	haService HAService
}

// NewService creates a new report Service.
func NewService(goeClient GoeClient, haService HAService) *Service {
	return &Service{
		goeClient: goeClient,
		haService: haService,
	}
}

// GenerateReportData orchestrates the fetching and formatting of the charging data
// to be output by a formatter.
func (s *Service) GenerateReportData(monthFlag, fromMonthFlag, toMonthFlag string) (models.ReportData, error) {
	// Validation
	if monthFlag == "" && fromMonthFlag == "" && toMonthFlag == "" {
		monthFlag = s.getPreviousMonth()
	}

	if monthFlag != "" && (fromMonthFlag != "" || toMonthFlag != "") {
		return models.ReportData{}, fmt.Errorf("cannot use --month together with --from-month/--to-month. Use one or the other")
	}

	startOfPeriod, endOfPeriod, periodLabel, err := s.getTimeRange(monthFlag, fromMonthFlag, toMonthFlag)
	if err != nil {
		return models.ReportData{}, err
	}

	fromMs := startOfPeriod.UnixNano() / 1e6
	toMs := endOfPeriod.UnixNano() / 1e6

	// Step 1: Fetch the direct JSON endpoint
	responseData, err := s.goeClient.FetchChargingData(fromMs, toMs)
	if err != nil {
		return models.ReportData{}, fmt.Errorf("error fetching JSON charging data: %w", err)
	}

	serial := viper.GetString(config.KeySerial)
	licensePlate := viper.GetString(config.KeyLicensePlate)
	kwhPrice := viper.GetFloat64(config.KeyKwhPrice)
	haSensorID := viper.GetString(config.KeyHAMilageSensor)
	chipIdsFlag := viper.GetString(config.KeyChipIds)

	// Step 4: Filter and aggregate data
	var reportData models.ReportData
	reportData.PeriodLabel = periodLabel
	reportData.StartDate = startOfPeriod.Format("02.01.2006")
	reportData.EndDate = endOfPeriod.Format("02.01.2006")
	reportData.SerialNumber = serial
	reportData.LicensePlate = licensePlate
	reportData.KwhPrice = kwhPrice

	if s.haService != nil && haSensorID != "" {
		mileage, err := s.haService.GetSensorValue(haSensorID)
		if err == nil {
			reportData.Mileage = mileage
		}
	}

	sessions, totalEnergy, totalPrice, totalSessions := s.processLogs(responseData, chipIdsFlag, kwhPrice)
	reportData.Sessions = sessions
	reportData.TotalEnergy = totalEnergy
	reportData.TotalPrice = totalPrice
	reportData.TotalSessions = totalSessions

	return reportData, nil
}

// processLogs filters raw charging data by RFID and maps it into the models.SessionData struct.
func (s *Service) processLogs(data *goe.DirectJsonResp, chipIdsFlag string, kwhPrice float64) (sessions []models.SessionData, totalEnergy, totalPrice float64, totalSessions int) {
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

		sessions = append(sessions, models.SessionData{
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

// getTimeRange parses the month flags and returns the start and end of the period along with a label.
// It returns an error if the flags are invalid.
func (s *Service) getTimeRange(monthFlag, fromMonthFlag, toMonthFlag string) (startOfPeriod, endOfPeriod time.Time, periodLabel string, err error) {
	if monthFlag != "" {
		// Single month mode (backward compatible)
		targetDate, parseErr := time.Parse("01-2006", monthFlag)
		if parseErr != nil {
			return time.Time{}, time.Time{}, "", fmt.Errorf("invalid date format for --month. Please use MM-YYYY (e.g. 02-2026)")
		}
		startOfPeriod = time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		endOfPeriod = startOfPeriod.AddDate(0, 1, 0).Add(-time.Nanosecond)
		periodLabel = monthFlag
	} else {
		// Multi-month mode
		fromDate, parseErr := time.Parse("01-2006", fromMonthFlag)
		if parseErr != nil {
			return time.Time{}, time.Time{}, "", fmt.Errorf("invalid date format for --from-month. Please use MM-YYYY (e.g. 02-2026)")
		}
		toDate, parseErr := time.Parse("01-2006", toMonthFlag)
		if parseErr != nil {
			return time.Time{}, time.Time{}, "", fmt.Errorf("invalid date format for --to-month. Please use MM-YYYY (e.g. 02-2026)")
		}

		if toDate.Before(fromDate) {
			return time.Time{}, time.Time{}, "", fmt.Errorf("--to-month must be equal to or after --from-month")
		}

		startOfPeriod = time.Date(fromDate.Year(), fromDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		endOfMonth := time.Date(toDate.Year(), toDate.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0).Add(-time.Nanosecond)
		endOfPeriod = endOfMonth
		periodLabel = fmt.Sprintf("%s_to_%s", fromMonthFlag, toMonthFlag)
	}
	return startOfPeriod, endOfPeriod, periodLabel, nil
}

// getPreviousMonth returns the MM-YYYY string of the month before the current time.
func (s *Service) getPreviousMonth() string {
	now := time.Now()
	// Use the 1st of the current month to avoid day overflow when subtracting a month
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	prevMonth := firstOfMonth.AddDate(0, -1, 0)
	return prevMonth.Format("01-2006")
}
