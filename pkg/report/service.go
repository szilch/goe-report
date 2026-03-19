package report

import (
	"fmt"
	"strings"
	"time"

	"echarge-report/pkg/models"
	"echarge-report/pkg/wallbox"
)

// WallboxAdapter defines the interface for fetching charging data from a wallbox.
type WallboxAdapter interface {
	FetchChargingData(fromMs, toMs int64) (*wallbox.ChargingResponse, error)
}

// CarInfoProvider defines the interface for retrieving car mileage information.
type CarInfoProvider interface {
	GetMileage() (int, error)
	GetMileageAt(t time.Time) (int, error)
}

// Config holds the configuration values required by Service to generate reports.
type Config struct {
	SerialNumber string
	LicensePlate string
	KwhPrice     float64
	ChipIDs      string
}

// Service generates charging reports by combining wallbox data with configuration.
type Service struct {
	wallboxAdapter  WallboxAdapter
	carInfoProvider CarInfoProvider
	cfg             Config
}

// NewService creates a new Service with the given wallbox adapter, optional car
// info provider, and report configuration.
func NewService(wallboxAdapter WallboxAdapter, carInfoProvider CarInfoProvider, cfg Config) *Service {
	return &Service{
		wallboxAdapter:  wallboxAdapter,
		carInfoProvider: carInfoProvider,
		cfg:             cfg,
	}
}

// GenerateReportData fetches charging sessions and assembles a ReportData for
// the given time period. monthFlag or fromMonthFlag/toMonthFlag must be set;
// if none are provided, the previous month is used.
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
		return models.ReportData{}, fmt.Errorf("fetch charging data: %w", err)
	}

	reportData := models.ReportData{
		PeriodLabel:  periodLabel,
		StartDate:    startOfPeriod,
		EndDate:      endOfPeriod,
		SerialNumber: s.cfg.SerialNumber,
		LicensePlate: s.cfg.LicensePlate,
		KwhPrice:     s.cfg.KwhPrice,
	}

	if s.carInfoProvider != nil {
		mileage, err := s.carInfoProvider.GetMileage()
		if err == nil {
			reportData.Mileage = mileage
		}

		mileageAtEnd, err := s.carInfoProvider.GetMileageAt(endOfPeriod)
		if err == nil {
			reportData.MileageAtEnd = mileageAtEnd
		}
	}

	sessions, totalEnergy, totalPrice, totalSessions := s.processLogs(responseData, s.cfg.ChipIDs, s.cfg.KwhPrice)
	reportData.Sessions = sessions
	reportData.TotalEnergy = totalEnergy
	reportData.TotalPrice = totalPrice
	reportData.TotalSessions = totalSessions

	return reportData, nil
}

func (s *Service) processLogs(data *wallbox.ChargingResponse, chipIDs string, kwhPrice float64) (sessions []models.SessionData, totalEnergy, totalPrice float64, totalSessions int) {
	for _, session := range data.Data {
		var idChipStr string
		if session.IdChip != nil {
			idChipStr = fmt.Sprintf("%v", session.IdChip)
		}

		matched := chipIDs == ""
		if !matched {
			for _, vid := range strings.Split(chipIDs, ",") {
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
	loc, loadErr := time.LoadLocation("Europe/Berlin")
	if loadErr != nil {
		// Fall back to UTC if the timezone database is unavailable.
		loc = time.UTC
	}

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
		endOfPeriod = time.Date(toDate.Year(), toDate.Month(), 1, 0, 0, 0, 0, loc).AddDate(0, 1, 0).Add(-time.Nanosecond)
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
