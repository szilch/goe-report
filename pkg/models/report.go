package models

import "time"

type SessionData struct {
	StartDate time.Time
	EndDate   time.Time
	Duration  string
	Energy    float64
	Price     float64
	RFID      string
}

type ReportData struct {
	PeriodLabel   string
	StartDate     time.Time
	EndDate       time.Time
	SerialNumber  string
	LicensePlate  string
	Mileage       int // Current mileage
	MileageAtEnd  int // Mileage at the end of the report period
	KwhPrice      float64
	TotalSessions int
	TotalEnergy   float64
	TotalPrice    float64
	Sessions      []SessionData
}
