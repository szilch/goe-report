package report

import (
	"fmt"
	"goe-report/pkg/config"
	"goe-report/pkg/goe"
	"goe-report/pkg/homeassistant"
	"goe-report/pkg/models"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Service provides functionality for generating charging reports.
type Service struct {
	goeClient *goe.Client
	haService *homeassistant.Service
}

// NewService creates a new report Service.
func NewService(goeClient *goe.Client, haService *homeassistant.Service) *Service {
	return &Service{
		goeClient: goeClient,
		haService: haService,
	}
}

// GenerateReportData orchestrates the fetching and formatting of the charging data
// to be output by a formatter.
func (s *Service) GenerateReportData(startOfPeriod, endOfPeriod time.Time, periodLabel string) (models.ReportData, error) {
	fromMs := startOfPeriod.UnixNano() / 1e6
	toMs := endOfPeriod.UnixNano() / 1e6

	// Step 1 & 2: Get ticket from the API
	ticket, err := s.goeClient.GetApiTicket()
	if err != nil {
		return models.ReportData{}, fmt.Errorf("error fetching API ticket: %w", err)
	}

	// Step 3: Fetch the direct JSON endpoint
	responseData, err := s.goeClient.FetchChargingData(ticket, fromMs, toMs)
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
	reportData.MonthName = periodLabel
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
