package formatter

import (
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var p = message.NewPrinter(language.German)

// FormatKWhPrice formats a float64 price as a string with 4 decimals, a comma separator, and the € symbol.
func FormatKWhPrice(price float64) string {
	return p.Sprintf("%.4f €", price)
}

// FormatPrice formats a float64 price as a string with 2 decimals, a comma separator, and the € symbol.
func FormatPrice(price float64) string {
	return p.Sprintf("%.2f €", price)
}

// SessionData represents a single charging session.
type SessionData struct {
	Date     string
	Duration string
	Energy   float64
	Price    float64
	RFID     string
}

// ReportData holds the aggregated data for the report.
type ReportData struct {
	MonthName     string
	StartDate     string
	EndDate       string
	SerialNumber  string
	LicensePlate  string
	Mileage       string // Mileage from Home Assistant (or "unknown")
	KwhPrice      float64
	TotalSessions int
	TotalEnergy   float64
	TotalPrice    float64
	Sessions      []SessionData
}

// Formatter defines the interface for different report output formats.
type Formatter interface {
	Format(data ReportData) error
}
