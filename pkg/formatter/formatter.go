package formatter

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
	SerialNumber  string
	LicensePlate  string
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
