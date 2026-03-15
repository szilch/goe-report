package formatter

import (
	"echarge-report/pkg/models"
	"fmt"
	"os"
	"text/tabwriter"
)

type TerminalFormatter struct{}

func NewTerminalFormatter() *TerminalFormatter {
	return &TerminalFormatter{}
}

func (f *TerminalFormatter) Format(data models.ReportData) error {
	fmt.Printf("\nLadehistorie für Wallbox\n")

	licPlate := data.LicensePlate
	if licPlate == "" {
		licPlate = "Keines hinterlegt"
	}

	fmt.Println("\nAbrechnungsdaten")
	fmt.Printf("Kfz-Kennzeichen: \t%s\n", licPlate)
	fmt.Printf("Kilometerstand:  \t%s\n", data.Mileage)
	fmt.Printf("Zeitraum:        \t%s - %s\n", data.StartDate.Format("02.01.2006"), data.EndDate.Format("02.01.2006"))
	fmt.Printf("Preis/kWh:       \t%s\n\n", FormatKWhPrice(data.KwhPrice))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintf(w, "Start\tEnde\tDauer\t%15s\t%9s\n", "Lademenge (kWh)", "Preis (€)")
	fmt.Fprintf(w, "-------------------\t-------------------\t--------\t%15s\t%9s\n", "---------------", "---------")

	for _, session := range data.Sessions {
		energyStr := fmt.Sprintf("%.2f kWh", session.Energy)
		priceStr := FormatPrice(session.Price)
		fmt.Fprintf(w, "%s\t%s\t%s\t%15s\t%9s\n", session.StartDate.Format("02.01.2006 15:04"), session.EndDate.Format("02.01.2006 15:04"), session.Duration, energyStr, priceStr)
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
