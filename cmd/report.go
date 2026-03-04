package cmd

import (
	"encoding/json"
	"fmt"
	"goe-report/pkg/formatter"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var chipIdsFlag string
var monthFlag string

// Struct matching the expected JSON response from the direct_json endpoint
type DirectJsonResp struct {
	Data []struct {
		IdChip       interface{} `json:"id_chip"`
		IdChipName   string      `json:"id_chip_name"`
		Start        string      `json:"start"`
		SecondsTotal string      `json:"seconds_total"`
		Energy       float64     `json:"energy"` // Assuming this is kWh based on sample
	} `json:"data"`
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a charging report for a specific RFID and month",
	Long:  `Fetches the charging history from the go-e Cloud API using the direct JSON endpoint and filters it by the provided RFID (or RFID Group) and month (in MM-YYYY format).`,
	Run: func(cmd *cobra.Command, args []string) {
		token := viper.GetString("token")
		serial := viper.GetString("serial")

		if token == "" || serial == "" {
			fmt.Println("Fehler: Token und Seriennummer müssen gesetzt sein.")
			fmt.Println("Nutze 'goe-report token set <token>' und 'goe-report serial set <serial>'.")
			os.Exit(1)
		}

		// Validation
		if monthFlag == "" {
			fmt.Println("Fehler: Der Parameter --month ist erforderlich (Format: MM-YYYY).")
			os.Exit(1)
		}

		// Parse target month
		targetDate, err := time.Parse("01-2006", monthFlag)
		if err != nil {
			fmt.Println("Fehler: Ungültiges Datumsformat für --month. Bitte MM-YYYY verwenden (z.B. 02-2026).")
			os.Exit(1)
		}

		// Calculate start and end of the month in milliseconds (UTC)
		startOfMonth := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

		fromMs := startOfMonth.UnixNano() / 1e6
		toMs := endOfMonth.UnixNano() / 1e6

		fmt.Printf("Frage Ladehistorie für Wallbox %s ab...\n", serial)

		// Step 1: Get the ticket DLL link to extract the e= parameter
		dllReqUrl := fmt.Sprintf("https://%s.api.v3.go-e.io/api/status?token=%s&filter=dll", serial, token)
		resp, err := http.Get(dllReqUrl)
		if err != nil {
			fmt.Printf("Fehler beim Abrufen des API-Tickets (Schritt 1): %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Fehler beim Lesen der Antwort: %v\n", err)
			os.Exit(1)
		}

		var dllResp struct {
			Dll string `json:"dll"`
		}
		if err := json.Unmarshal(body, &dllResp); err != nil {
			fmt.Printf("Fehler beim Parsen der API-Antwort: %v\n", err)
			os.Exit(1)
		}

		if dllResp.Dll == "" {
			fmt.Println("Fehler: Konnte kein Ticket von der API erhalten.")
			os.Exit(1)
		}

		// Step 2: Extract Ticket from DLL string
		parsedUrl, err := url.Parse(dllResp.Dll)
		if err != nil {
			fmt.Printf("Fehler beim Parsen der URL: %v\n", err)
			os.Exit(1)
		}
		ticket := parsedUrl.Query().Get("e")
		if ticket == "" {
			fmt.Println("Fehler: Ticket konnte nicht extrahiert werden.")
			os.Exit(1)
		}

		// Step 3: Fetch the direct JSON endpoint
		jsonUrl := fmt.Sprintf("https://data.v3.go-e.io/api/v1/direct_json?e=%s&from=%d&to=%d&timezone=Europe/Berlin", ticket, fromMs, toMs)

		jsonResp, err := http.Get(jsonUrl)
		if err != nil {
			fmt.Printf("Fehler beim Abrufen der JSON Ladedaten: %v\n", err)
			os.Exit(1)
		}
		defer jsonResp.Body.Close()

		jsonBody, err := io.ReadAll(jsonResp.Body)
		if err != nil {
			fmt.Printf("Fehler beim Lesen der JSON-Daten: %v\n", err)
			os.Exit(1)
		}

		var responseData DirectJsonResp
		if err := json.Unmarshal(jsonBody, &responseData); err != nil {
			fmt.Printf("Fehler beim Parsen der Ladedaten (JSON): %v\n", err)
			os.Exit(1)
		}

		// Step 4: Filter and aggregate data
		var reportData formatter.ReportData
		reportData.MonthName = monthFlag
		reportData.SerialNumber = serial
		reportData.LicensePlate = viper.GetString("licenseplate")

		for _, session := range responseData.Data {
			// Convert IdChip to string safely
			var idChipStr string
			if session.IdChip != nil {
				idChipStr = fmt.Sprintf("%v", session.IdChip)
			}

			// Check RFID matching (ID or Name depending on user preference, we test both for flexibility)
			// According to API, sometimes id_chip is empty but id_chip_uid is used, we simplify based on standard responses.
			// Let's check both id_chip and id_chip_name against the requested flags

			// Check RFID matching against chipIdsFlag
			matched := false
			if chipIdsFlag == "" {
				// If no chip IDs filter is provided, include all
				matched = true
			} else {
				// Check if the current ID or Name is in the provided list
				validIds := strings.Split(chipIdsFlag, ",")
				for _, vid := range validIds {
					v := strings.TrimSpace(vid)
					if idChipStr == v || session.IdChipName == v {
						matched = true
						break
					}
				}
			}

			if !matched {
				continue
			}

			reportData.TotalEnergy += session.Energy
			reportData.TotalSessions++

			reportData.Sessions = append(reportData.Sessions, formatter.SessionData{
				Date:     session.Start,
				Duration: session.SecondsTotal,
				Energy:   session.Energy,
				RFID:     idChipStr,
			})
		}

		// Step 5: Execute the corresponding formatter
		var frm formatter.Formatter
		if pdfFlag {
			filename := fmt.Sprintf("goe_report_%s.pdf", monthFlag)
			if chipIdsFlag != "" {
				safeIds := strings.ReplaceAll(chipIdsFlag, ",", "_")
				filename = fmt.Sprintf("goe_report_%s_%s.pdf", monthFlag, safeIds)
			}
			frm = formatter.NewPDFFormatter(filename)
		} else {
			frm = formatter.NewTerminalFormatter()
		}

		if err := frm.Format(reportData); err != nil {
			fmt.Printf("Fehler bei der Reportausgabe: %v\n", err)
			os.Exit(1)
		}
	},
}

var pdfFlag bool

func init() {
	reportCmd.Flags().StringVar(&chipIdsFlag, "chipIds", "", "Optional. Kommaseparierte Liste von Chip-IDs zur Filterung (z.B. 12345,67890)")
	reportCmd.Flags().StringVar(&monthFlag, "month", "", "Zwingend erforderlich. Monat im Format MM-YYYY (z.B. 02-2026)")
	reportCmd.Flags().BoolVar(&pdfFlag, "pdf", false, "Gibt den Report als PDF-Datei aus.")

	rootCmd.AddCommand(reportCmd)
}
