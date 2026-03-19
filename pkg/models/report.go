package models

import "time"

// SessionData contains the specific details for a single charging session.
type SessionData struct {
	StartDate time.Time
	EndDate   time.Time
	Duration  string
	Energy    float64
	Price     float64
	RFID      string
}

// ReportData holds the complete aggregated data for a charging report, including
// overall statistics and a breakdown of individual sessions.
type ReportData struct {
	PeriodLabel   string
	StartDate     time.Time
	EndDate       time.Time
	SerialNumber  string
	LicensePlate  string
	Driver        string
	Mileage       int // Current mileage
	MileageAtEnd  int // Mileage at the end of the report period
	HasMileage    bool // Indicates if mileage should be displayed
	KwhPrice      float64
	TotalSessions int
	TotalEnergy   float64
	TotalPrice    float64
	Sessions      []SessionData
}
