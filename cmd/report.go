package cmd

import (
	"encoding/json"
	"fmt"
	"goe-report/pkg/config"
	"goe-report/pkg/formatter"
	"goe-report/pkg/ha"
	"goe-report/pkg/mail"
	"goe-report/pkg/pdfmerge"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
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
		token := viper.GetString(config.KeyToken)
		serial := viper.GetString(config.KeySerial)

		if token == "" || serial == "" {
			color.Red("Error: Token and serial number must be set.")
			color.Red("Use 'goe-report config-set goe_token <token>' and 'goe-report config-set goe_serial <serial>'.")
			os.Exit(1)
		}

		// --attach-pdfs requires --pdf
		if attachPdfsFlag && !pdfFlag {
			color.Red("Error: --attach-pdfs requires --pdf to be set.")
			os.Exit(1)
		}

		// --send-mail requires --pdf
		if sendMailFlag && !pdfFlag {
			color.Red("Error: --send-mail requires --pdf to be set.")
			os.Exit(1)
		}

		// Validation
		if monthFlag == "" {
			color.Red("Error: The --month parameter is required (format: MM-YYYY).")
			os.Exit(1)
		}

		// Parse target month
		targetDate, err := time.Parse("01-2006", monthFlag)
		if err != nil {
			color.Red("Error: Invalid date format for --month. Please use MM-YYYY (e.g. 02-2026).")
			os.Exit(1)
		}

		// Calculate start and end of the month in milliseconds (UTC)
		startOfMonth := time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

		fromMs := startOfMonth.UnixNano() / 1e6
		toMs := endOfMonth.UnixNano() / 1e6

		color.Blue("Fetching charging history for wallbox %s...", serial)

		// Step 1: Get the ticket DLL link to extract the e= parameter
		dllReqUrl := fmt.Sprintf("https://%s.api.v3.go-e.io/api/status?token=%s&filter=dll", serial, token)
		resp, err := http.Get(dllReqUrl)
		if err != nil {
			color.Red("Error fetching API ticket (step 1): %v", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			color.Red("Error reading response: %v", err)
			os.Exit(1)
		}

		var dllResp struct {
			Dll string `json:"dll"`
		}
		if err := json.Unmarshal(body, &dllResp); err != nil {
			color.Red("Error parsing API response: %v", err)
			os.Exit(1)
		}

		if dllResp.Dll == "" {
			color.Red("Error: Could not obtain a ticket from the API.")
			os.Exit(1)
		}

		// Step 2: Extract Ticket from DLL string
		parsedUrl, err := url.Parse(dllResp.Dll)
		if err != nil {
			color.Red("Error parsing URL: %v", err)
			os.Exit(1)
		}
		ticket := parsedUrl.Query().Get("e")
		if ticket == "" {
			color.Red("Error: Could not extract ticket from URL.")
			os.Exit(1)
		}

		// Step 3: Fetch the direct JSON endpoint
		jsonUrl := fmt.Sprintf("https://data.v3.go-e.io/api/v1/direct_json?e=%s&from=%d&to=%d&timezone=Europe/Berlin", ticket, fromMs, toMs)

		jsonResp, err := http.Get(jsonUrl)
		if err != nil {
			color.Red("Error fetching JSON charging data: %v", err)
			os.Exit(1)
		}
		defer jsonResp.Body.Close()

		jsonBody, err := io.ReadAll(jsonResp.Body)
		if err != nil {
			color.Red("Error reading JSON data: %v", err)
			os.Exit(1)
		}

		var responseData DirectJsonResp
		if err := json.Unmarshal(jsonBody, &responseData); err != nil {
			color.Red("Error parsing charging data (JSON): %v", err)
			os.Exit(1)
		}

		// Step 4: Filter and aggregate data
		var reportData formatter.ReportData
		reportData.MonthName = monthFlag
		reportData.StartDate = startOfMonth.Format("02.01.2006")
		reportData.EndDate = endOfMonth.Format("02.01.2006")
		reportData.SerialNumber = serial
		reportData.LicensePlate = viper.GetString(config.KeyLicensePlate)

		// Fetch mileage from Home Assistant
		color.Blue("Fetching mileage from Home Assistant...")
		haService := ha.NewService(viper.GetString(config.KeyHAAPI), viper.GetString(config.KeyHAToken))
		mileage, err := haService.GetSensorValue(viper.GetString(config.KeyHAMilageSensor))
		if err != nil {
			color.Yellow("Warning: Could not fetch Home Assistant mileage: %v", err)
		}
		reportData.Mileage = mileage

		kwhPrice := viper.GetFloat64(config.KeyKwhPrice)
		reportData.KwhPrice = kwhPrice

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

			sessionPrice := session.Energy * kwhPrice

			reportData.TotalEnergy += session.Energy
			reportData.TotalSessions++
			reportData.TotalPrice += sessionPrice

			reportData.Sessions = append(reportData.Sessions, formatter.SessionData{
				Date:     session.Start,
				Duration: session.SecondsTotal,
				Energy:   session.Energy,
				Price:    sessionPrice,
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
			color.Red("Error generating report output: %v", err)
			os.Exit(1)
		}

		// Attach PDFs from ~/.goe-report/ if requested
		if pdfFlag && attachPdfsFlag {
			home, err := os.UserHomeDir()
			if err != nil {
				color.Red("Error determining home directory: %v", err)
				os.Exit(1)
			}
			configDir := filepath.Join(home, ".goe-report")

			// Collect all PDFs from the config directory, excluding the report we just generated.
			matches, err := filepath.Glob(filepath.Join(configDir, "*.pdf"))
			if err != nil {
				color.Red("Error scanning for attachment PDFs: %v", err)
				os.Exit(1)
			}

			// Determine the absolute path of the generated report file.
			reportFile := fmt.Sprintf("goe_report_%s.pdf", monthFlag)
			if chipIdsFlag != "" {
				safeIds := strings.ReplaceAll(chipIdsFlag, ",", "_")
				reportFile = fmt.Sprintf("goe_report_%s_%s.pdf", monthFlag, safeIds)
			}
			reportAbs, _ := filepath.Abs(reportFile)

			var attachments []string
			for _, m := range matches {
				abs, _ := filepath.Abs(m)
				if abs == reportAbs {
					continue // skip the report itself
				}
				attachments = append(attachments, m)
			}

			if len(attachments) == 0 {
				color.Yellow("Warning: --attach-pdfs set but no PDF files found in %s.", configDir)
			} else {
				color.Blue("Attaching %d PDF(s) from %s...", len(attachments), configDir)
				if err := pdfmerge.Merge(reportFile, attachments); err != nil {
					color.Red("Error attaching PDFs: %v", err)
					os.Exit(1)
				}
				color.Blue("PDFs attached successfully.")
			}
		}

		// Send email if requested
		if sendMailFlag {
			color.Blue("Preparing to send email...")

			// Get mail_to addresses
			toRaw := viper.GetString(config.KeyMailTo)
			if toRaw == "" {
				color.Red("Error: Cannot send email because 'mail_to' is not configured. Use 'goe-report config-set mail_to ...'")
				os.Exit(1)
			}

			var recipients []string
			for _, r := range strings.Split(toRaw, ",") {
				if trimmed := strings.TrimSpace(r); trimmed != "" {
					recipients = append(recipients, trimmed)
				}
			}

			if len(recipients) == 0 {
				color.Red("Error: No valid recipient addresses found in 'mail_to' configuration.")
				os.Exit(1)
			}

			// Determine filename (same logic as above)
			reportFile := fmt.Sprintf("goe_report_%s.pdf", monthFlag)
			if chipIdsFlag != "" {
				safeIds := strings.ReplaceAll(chipIdsFlag, ",", "_")
				reportFile = fmt.Sprintf("goe_report_%s_%s.pdf", monthFlag, safeIds)
			}

			// Read PDF file data
			pdfData, err := os.ReadFile(reportFile)
			if err != nil {
				color.Red("Error reading generated PDF for email attachment: %v", err)
				os.Exit(1)
			}

			// Mail configuration
			cfg := mail.Config{
				Host:     viper.GetString(config.KeyMailHost),
				Port:     viper.GetInt(config.KeyMailPort),
				Username: viper.GetString(config.KeyMailUsername),
				Password: viper.GetString(config.KeyMailPassword),
				From:     viper.GetString(config.KeyMailFrom),
			}
			mailer := mail.NewMailService(cfg)

			subject := fmt.Sprintf("Ladebericht - %s (%s)", viper.GetString(config.KeyLicensePlate), monthFlag)
			body := fmt.Sprintf("Hallo,\n\nangehängt findest du den Ladebericht für das Kennzeichen %s für den Zeitraum %s.\n\nViele Grüße,\ngoe-report", viper.GetString(config.KeyLicensePlate), monthFlag)
			attachment := mail.Attachment{
				Name: reportFile,
				Data: pdfData,
			}

			color.Blue("Sending email to %v...", recipients)
			if err := mailer.Send(recipients, subject, body, attachment); err != nil {
				color.Red("Error sending email: %v", err)
				os.Exit(1)
			}
			color.Green("Email sent successfully.")
		}
	},
}

var pdfFlag bool
var attachPdfsFlag bool
var sendMailFlag bool

func init() {
	reportCmd.Flags().StringVar(&chipIdsFlag, "chipIds", "", "Optional. Comma-separated list of chip IDs to filter by (e.g. 12345,67890)")
	reportCmd.Flags().StringVar(&monthFlag, "month", "", "Required. Month in MM-YYYY format (e.g. 02-2026)")
	reportCmd.Flags().BoolVar(&pdfFlag, "pdf", false, "Export the report as a PDF file.")
	reportCmd.Flags().BoolVar(&attachPdfsFlag, "attach-pdfs", false, "Attach all PDF files from ~/.goe-report/ to the generated report PDF. Requires --pdf.")
	reportCmd.Flags().BoolVar(&sendMailFlag, "send-mail", false, "Send the generated PDF via email. Requires --pdf and configured mail settings (-h for details).")

	rootCmd.AddCommand(reportCmd)
}
