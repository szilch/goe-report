package formatter

import (
	"goe-report/pkg/models"

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


// Formatter defines the interface for different report output formats.
type Formatter interface {
	Format(data models.ReportData) error
}
