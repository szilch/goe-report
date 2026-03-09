package cmd

import (
	"fmt"
	"goe-report/pkg/config"
	"goe-report/pkg/formatter"
	"goe-report/pkg/goe"
	"goe-report/pkg/homeassistant"
	"goe-report/pkg/mail"
	"goe-report/pkg/pdfmerge"
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
var fromMonthFlag string
var toMonthFlag string

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
	Short: "Generate a charging report for a specific RFID and month or date range",
	Long:  `Fetches the charging history from the go-e Cloud API using the direct JSON endpoint and filters it by the provided RFID (or RFID Group) and month (in MM-YYYY format) or a date range (using --from-month and --to-month).`,
	Run: func(cmd *cobra.Command, args []string) {
		token := viper.GetString(config.KeyToken)
		serial := viper.GetString(config.KeySerial)
		localApiUrl := viper.GetString(config.KeyLocalApiUrl)

		if chipIdsFlag == "" {
			chipIdsFlag = viper.GetString(config.KeyChipIds)
		}

		if serial == "" {
			color.Red("Error: Serial number must be set.")
			color.Red("Use 'goe-report config-set goe_serial <serial>'.")
			os.Exit(1)
		}

		if token == "" && localApiUrl == "" {
			color.Red("Error: Either a Cloud API Token or a Local API URL must be configured.")
			color.Red("Use 'goe-report config-set goe_token <token>' or 'goe-report config-set goe_localApiUrl http://<ip>'.")
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
		if monthFlag == "" && (fromMonthFlag == "" || toMonthFlag == "") {
			color.Red("Error: Either --month or both --from-month and --to-month parameters are required (format: MM-YYYY).")
			os.Exit(1)
		}

		if monthFlag != "" && (fromMonthFlag != "" || toMonthFlag != "") {
			color.Red("Error: Cannot use --month together with --from-month/--to-month. Use one or the other.")
			os.Exit(1)
		}

		var startOfPeriod, endOfPeriod time.Time
		var periodLabel string

		if monthFlag != "" {
			// Single month mode (backward compatible)
			targetDate, err := time.Parse("01-2006", monthFlag)
			if err != nil {
				color.Red("Error: Invalid date format for --month. Please use MM-YYYY (e.g. 02-2026).")
				os.Exit(1)
			}
			startOfPeriod = time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, time.UTC)
			endOfPeriod = startOfPeriod.AddDate(0, 1, 0).Add(-time.Nanosecond)
			periodLabel = monthFlag
		} else {
			// Multi-month mode
			fromDate, err := time.Parse("01-2006", fromMonthFlag)
			if err != nil {
				color.Red("Error: Invalid date format for --from-month. Please use MM-YYYY (e.g. 02-2026).")
				os.Exit(1)
			}
			toDate, err := time.Parse("01-2006", toMonthFlag)
			if err != nil {
				color.Red("Error: Invalid date format for --to-month. Please use MM-YYYY (e.g. 02-2026).")
				os.Exit(1)
			}

			if toDate.Before(fromDate) {
				color.Red("Error: --to-month must be equal to or after --from-month.")
				os.Exit(1)
			}

			startOfPeriod = time.Date(fromDate.Year(), fromDate.Month(), 1, 0, 0, 0, 0, time.UTC)
			endOfMonth := time.Date(toDate.Year(), toDate.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0).Add(-time.Nanosecond)
			endOfPeriod = endOfMonth
			periodLabel = fmt.Sprintf("%s_to_%s", fromMonthFlag, toMonthFlag)
		}

		fromMs := startOfPeriod.UnixNano() / 1e6
		toMs := endOfPeriod.UnixNano() / 1e6

		color.Blue("Fetching charging history for wallbox %s...", serial)

		client := goe.NewClient(serial, token, localApiUrl)

		// Step 1 & 2: Get ticket from the API
		ticket, err := client.GetApiTicket()
		if err != nil {
			color.Red("Error fetching API ticket: %v", err)
			os.Exit(1)
		}

		// Step 3: Fetch the direct JSON endpoint
		responseData, err := client.FetchChargingData(ticket, fromMs, toMs)
		if err != nil {
			color.Red("Error fetching JSON charging data: %v", err)
			os.Exit(1)
		}

		// Step 4: Filter and aggregate data
		var reportData formatter.ReportData
		reportData.MonthName = periodLabel
		reportData.StartDate = startOfPeriod.Format("02.01.2006")
		reportData.EndDate = endOfPeriod.Format("02.01.2006")
		reportData.SerialNumber = serial
		reportData.LicensePlate = viper.GetString(config.KeyLicensePlate)

		// Fetch mileage from Home Assistant
		color.Blue("Fetching mileage from Home Assistant...")
		haService := homeassistant.NewService(viper.GetString(config.KeyHAAPI), viper.GetString(config.KeyHAToken))
		mileage, err := haService.GetSensorValue(viper.GetString(config.KeyHAMilageSensor))
		if err != nil {
			color.Yellow("Warning: Could not fetch Home Assistant mileage: %v", err)
		}
		reportData.Mileage = mileage

		kwhPrice := viper.GetFloat64(config.KeyKwhPrice)
		reportData.KwhPrice = kwhPrice

		sessions, totalEnergy, totalPrice, totalSessions := goe.ProcessLogs(responseData, chipIdsFlag, kwhPrice)
		reportData.Sessions = sessions
		reportData.TotalEnergy = totalEnergy
		reportData.TotalPrice = totalPrice
		reportData.TotalSessions = totalSessions

		// Step 5: Execute the corresponding formatter
		var frm formatter.Formatter
		if pdfFlag {
			filename := fmt.Sprintf("goe_report_%s.pdf", periodLabel)
			if chipIdsFlag != "" {
				safeIds := strings.ReplaceAll(chipIdsFlag, ",", "_")
				filename = fmt.Sprintf("goe_report_%s_%s.pdf", periodLabel, safeIds)
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
			reportFile := fmt.Sprintf("goe_report_%s.pdf", periodLabel)
			if chipIdsFlag != "" {
				safeIds := strings.ReplaceAll(chipIdsFlag, ",", "_")
				reportFile = fmt.Sprintf("goe_report_%s_%s.pdf", periodLabel, safeIds)
			}
			if err := attachPDFs(reportFile); err != nil {
				color.Red("%v", err)
				os.Exit(1)
			}
		}

		// Send email if requested
		if sendMailFlag {
			reportFile := fmt.Sprintf("goe_report_%s.pdf", periodLabel)
			if chipIdsFlag != "" {
				safeIds := strings.ReplaceAll(chipIdsFlag, ",", "_")
				reportFile = fmt.Sprintf("goe_report_%s_%s.pdf", periodLabel, safeIds)
			}
			if err := sendReportEmail(reportFile, periodLabel, viper.GetString(config.KeyLicensePlate)); err != nil {
				color.Red("%v", err)
				os.Exit(1)
			}
		}
	},
}

var pdfFlag bool
var attachPdfsFlag bool
var sendMailFlag bool

func sendReportEmail(reportFile, monthFlag, licensePlate string) error {
	color.Blue("Preparing to send email...")

	toRaw := viper.GetString(config.KeyMailTo)
	if toRaw == "" {
		return fmt.Errorf("cannot send email because 'mail_to' is not configured")
	}

	var recipients []string
	for _, r := range strings.Split(toRaw, ",") {
		if trimmed := strings.TrimSpace(r); trimmed != "" {
			recipients = append(recipients, trimmed)
		}
	}

	if len(recipients) == 0 {
		return fmt.Errorf("no valid recipient addresses found in 'mail_to' configuration")
	}

	pdfData, err := os.ReadFile(reportFile)
	if err != nil {
		return fmt.Errorf("error reading generated PDF for email attachment: %w", err)
	}

	// Mail configuration
	cfg := mail.Config{
		Host:     viper.GetString(config.KeyMailHost),
		Port:     viper.GetInt(config.KeyMailPort),
		Username: viper.GetString(config.KeyMailUsername),
		Password: viper.GetString(config.KeyMailPassword),
		From:     viper.GetString(config.KeyMailFrom),
	}
	mailer := mail.NewService(cfg)

	subject := fmt.Sprintf("Ladebericht - %s (%s)", licensePlate, monthFlag)
	body := fmt.Sprintf("Hallo,\n\nangehängt findest du den Ladebericht für das Kennzeichen %s für den Zeitraum %s.\n\nViele Grüße,\ngoe-report", licensePlate, monthFlag)
	attachment := mail.Attachment{
		Name: reportFile,
		Data: pdfData,
	}

	color.Blue("Sending email to %v...", recipients)
	if err := mailer.Send(recipients, subject, body, attachment); err != nil {
		return fmt.Errorf("error sending email: %w", err)
	}
	color.Green("Email sent successfully.")
	return nil
}

func attachPDFs(reportFile string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error determining home directory: %w", err)
	}
	configDir := filepath.Join(home, config.ConfigDirName)

	matches, err := filepath.Glob(filepath.Join(configDir, "*.pdf"))
	if err != nil {
		return fmt.Errorf("error scanning for attachment PDFs: %w", err)
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
		return nil
	}

	color.Blue("Attaching %d PDF(s) from %s...", len(attachments), configDir)
	if err := pdfmerge.Merge(reportFile, attachments); err != nil {
		return fmt.Errorf("error attaching PDFs: %w", err)
	}
	color.Blue("PDFs attached successfully.")
	return nil
}

func init() {
	reportCmd.Flags().StringVar(&chipIdsFlag, "chipIds", "", "Optional. Comma-separated list of chip IDs to filter by (e.g. 12345,67890)")
	reportCmd.Flags().StringVar(&monthFlag, "month", "", "Month in MM-YYYY format (e.g. 02-2026). Use this for a single month report.")
	reportCmd.Flags().StringVar(&fromMonthFlag, "from-month", "", "Start month in MM-YYYY format (e.g. 01-2026). Use with --to-month for multi-month reports.")
	reportCmd.Flags().StringVar(&toMonthFlag, "to-month", "", "End month in MM-YYYY format (e.g. 03-2026). Use with --from-month for multi-month reports.")
	reportCmd.Flags().BoolVar(&pdfFlag, "pdf", false, "Export the report as a PDF file.")
	reportCmd.Flags().BoolVar(&attachPdfsFlag, "attach-pdfs", false, fmt.Sprintf("Attach all PDF files from ~/%s/ to the generated report PDF. Requires --pdf.", config.ConfigDirName))
	reportCmd.Flags().BoolVar(&sendMailFlag, "send-mail", false, "Send the generated PDF via email. Requires --pdf and configured mail settings (-h for details).")

	rootCmd.AddCommand(reportCmd)
}
