package report

import (
	"echarge-report/pkg/config"
	"echarge-report/pkg/models"
	"echarge-report/pkg/wallbox"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type WallboxAdapter interface {
	FetchChargingData(fromMs, toMs int64) (*wallbox.ChargingResponse, error)
}

type HAService interface {
	GetSensorValue(sensorID string) (string, error)
}

type Service struct {
	wallboxAdapter WallboxAdapter
	haService      HAService
}

func NewService(wallboxAdapter WallboxAdapter, haService HAService) *Service {
	return &Service{
		wallboxAdapter: wallboxAdapter,
		haService:      haService,
	}
}

func (s *Service) GenerateReportData(monthFlag, fromMonthFlag, toMonthFlag string) (models.ReportData, error) {
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

	responseData, err := s.wallboxAdapter.FetchChargingData(fromMs, toMs)
	if err != nil {
		return models.ReportData{}, fmt.Errorf("error fetching charging data: %w", err)
	}

	serial := viper.GetString(config.KeyWallboxGoeCloudSerial)
	licensePlate := viper.GetString(config.KeyLicensePlate)
	kwhPrice := viper.GetFloat64(config.KeyKwhPrice)
	haSensorID := viper.GetString(config.KeyHAMilageSensor)
	chipIdsFlag := viper.GetString(config.KeyWallboxChipIds)

	var reportData models.ReportData
	reportData.PeriodLabel = periodLabel
	reportData.StartDate = startOfPeriod
	reportData.EndDate = endOfPeriod
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

func (s *Service) processLogs(data *wallbox.ChargingResponse, chipIdsFlag string, kwhPrice float64) (sessions []models.SessionData, totalEnergy, totalPrice float64, totalSessions int) {
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
			Duration:  s.formatDuration(session.Duration),
			Energy:    session.Energy,
			Price:     sessionPrice,
			RFID:      idChipStr,
		})
	}

	return sessions, totalEnergy, totalPrice, totalSessions
}

func (s *Service) getTimeRange(monthFlag, fromMonthFlag, toMonthFlag string) (startOfPeriod, endOfPeriod time.Time, periodLabel string, err error) {
	loc, _ := time.LoadLocation("Europe/Berlin")

	if monthFlag != "" {
		targetDate, parseErr := time.Parse("01-2006", monthFlag)
		if parseErr != nil {
			return time.Time{}, time.Time{}, "", fmt.Errorf("invalid date format for --month. Please use MM-YYYY (e.g. 02-2026)")
		}
		startOfPeriod = time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, loc)
		endOfPeriod = startOfPeriod.AddDate(0, 1, 0).Add(-time.Nanosecond)
		periodLabel = monthFlag
	} else {
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

		startOfPeriod = time.Date(fromDate.Year(), fromDate.Month(), 1, 0, 0, 0, 0, loc)
		endOfMonth := time.Date(toDate.Year(), toDate.Month(), 1, 0, 0, 0, 0, loc).AddDate(0, 1, 0).Add(-time.Nanosecond)
		endOfPeriod = endOfMonth
		periodLabel = fmt.Sprintf("%s_to_%s", fromMonthFlag, toMonthFlag)
	}
	return startOfPeriod, endOfPeriod, periodLabel, nil
}

func (s *Service) getPreviousMonth() string {
	now := time.Now()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	prevMonth := firstOfMonth.AddDate(0, -1, 0)
	return prevMonth.Format("01-2006")
}

func (s *Service) formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	sec := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, sec)
}
