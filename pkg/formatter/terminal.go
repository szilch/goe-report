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
	fmt.Printf("\nLadehistorie für Wallbox\n")

	licPlate := data.LicensePlate
	if licPlate == "" {
		licPlate = "Keines hinterlegt"
	}

	fmt.Println("\nAbrechnungsdaten")
	fmt.Printf("Kfz-Kennzeichen: \t%s\n", licPlate)
	fmt.Printf("Kilometerstand:  \t%s\n", data.Mileage)
	fmt.Printf("Zeitraum:        \t%s - %s\n", data.StartDate, data.EndDate)
	fmt.Printf("Preis/kWh:       \t%s\n\n", FormatKWhPrice(data.KwhPrice))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "Start\tEnde\tDauer\t%15s\t%9s\n", "Lademenge (kWh)", "Preis (€)")
	fmt.Fprintf(w, "-------------------\t-------------------\t--------\t%15s\t%9s\n", "---------------", "---------")

	for _, session := range data.Sessions {
		energyStr := fmt.Sprintf("%.2f kWh", session.Energy)
		priceStr := FormatPrice(session.Price)
		fmt.Fprintf(w, "%s\t%s\t%s\t%15s\t%9s\n", session.StartDate, session.EndDate, session.Duration, energyStr, priceStr)
	}

	w.Flush()

	fmt.Println("----------------------------------------------------------------------------------")
	if data.TotalSessions == 0 {
		fmt.Println("Keine Ladevorgänge für diese Kriterien im gewünschten Zeitraum gefunden.")
	} else {
		fmt.Printf("Gesamte Ladevorgänge:\t%d\n", data.TotalSessions)
		fmt.Printf("Gesamte Energie:\t%.2f kWh\n", data.TotalEnergy)
		fmt.Printf("Gesamtpreis:\t\t%s\n", FormatPrice(data.TotalPrice))
	}

	return nil
}
