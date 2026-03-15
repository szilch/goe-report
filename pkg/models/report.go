package models

import "time"


// SessionData represents a single charging session.
type SessionData struct {
	StartDate time.Time
	EndDate   time.Time
	Duration  string
	Energy    float64
	Price     float64
	RFID      string
}

// ReportData holds the aggregated data for the report.
type ReportData struct {
	PeriodLabel   string
	StartDate     time.Time
	EndDate       time.Time
	SerialNumber  string
	LicensePlate  string
	Mileage       string // Mileage from Home Assistant (or "unknown")
	KwhPrice      float64
	TotalSessions int
	TotalEnergy   float64
	TotalPrice    float64
	Sessions      []SessionData
}
