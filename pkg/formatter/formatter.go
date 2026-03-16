package formatter

import (
	"echarge-report/pkg/models"
	"fmt"

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

// FormatMileage formats an int mileage as a string, returning "---" if the mileage is 0.
func FormatMileage(mileage int) string {
	if mileage == 0 {
		return "---"
	}
	return fmt.Sprintf("%d km", mileage)
}

// Formatter defines the interface for different report output formats.
type Formatter interface {
	Format(data models.ReportData) error
}
