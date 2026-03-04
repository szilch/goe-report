package formatter

import (
	"fmt"
	"os"
	"text/tabwriter"
)

// TerminalFormatter outputs the report to the console using tabwriter.
type TerminalFormatter struct{}

// NewTerminalFormatter creates a new TerminalFormatter.
func NewTerminalFormatter() *TerminalFormatter {
	return &TerminalFormatter{}
}

// Format prints the ReportData to os.Stdout in a tabulated format.
func (f *TerminalFormatter) Format(data ReportData) error {
	fmt.Printf("\nLadehistorie %s für Wallbox %s\n", data.MonthName, data.SerialNumber)

	licPlate := data.LicensePlate
	if licPlate == "" {
		licPlate = "Keines hinterlegt"
	}

	fmt.Println("\nAbrechnungsdaten")
	fmt.Printf("Kfz-Kennzeichen: \t%s\n", licPlate)
	fmt.Printf("Zeitraum: \t\t%s\n\n", data.MonthName)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "Datum\tDauer\tLademenge (kWh)\n")
	fmt.Fprintf(w, "-----\t-----\t---------------\n")

	for _, session := range data.Sessions {
		fmt.Fprintf(w, "%s\t%s\t%.2f\n", session.Date, session.Duration, session.Energy)
	}

	w.Flush()

	fmt.Println("---------------------------------------------------------")
	if data.TotalSessions == 0 {
		fmt.Println("Keine Ladevorgänge für diese Kriterien im gewünschten Zeitraum gefunden.")
	} else {
		fmt.Printf("Gesamte Ladevorgänge:\t%d\n", data.TotalSessions)
		fmt.Printf("Gesamte Energie:\t%.2f kWh\n", data.TotalEnergy)
	}

	return nil
}
