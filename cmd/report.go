package cmd

import (
	"fmt"
	"goe-report/pkg/config"
	"goe-report/pkg/formatter"
	"goe-report/pkg/goe"
	"goe-report/pkg/homeassistant"
	"goe-report/pkg/mail"
	"goe-report/pkg/pdfmerge"
	"goe-report/pkg/report"
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
		if monthFlag == "" && fromMonthFlag == "" && toMonthFlag == "" {
			monthFlag = getPreviousMonth()
			color.Blue("No month specified, using previous month: %s", monthFlag)
		}

		if monthFlag != "" && (fromMonthFlag != "" || toMonthFlag != "") {
			color.Red("Error: Cannot use --month together with --from-month/--to-month. Use one or the other.")
			os.Exit(1)
		}

		startOfPeriod, endOfPeriod, periodLabel, err := getTimeRange(monthFlag, fromMonthFlag, toMonthFlag)
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		color.Blue("Fetching charging history for wallbox %s...", serial)

		client := goe.NewClient()
		haService := homeassistant.NewService()
		reportSvc := report.NewService(client, haService)

		reportData, err := reportSvc.GenerateReportData(
			startOfPeriod,
			endOfPeriod,
			periodLabel,
		)
		if err != nil {
			color.Red("Error generating report data: %v", err)
			os.Exit(1)
		}

		// Step 5: Execute the corresponding formatter
		var frm formatter.Formatter
		var reportFilename string
		if pdfFlag {
			reportFilename = fmt.Sprintf("goe_report_%s.pdf", periodLabel)
			frm = formatter.NewPDFFormatter(reportFilename)
		} else {
			frm = formatter.NewTerminalFormatter()
		}

		if err := frm.Format(reportData); err != nil {
			color.Red("Error generating report output: %v", err)
			os.Exit(1)
		}

		// Attach PDFs from ~/.goe-report/ if requested
		if pdfFlag && attachPdfsFlag {
			if err := attachPDFs(reportFilename); err != nil {
				color.Red("%v", err)
				os.Exit(1)
			}
		}

		// Send email if requested
		if sendMailFlag {
			if err := sendReportEmail(reportFilename, periodLabel, viper.GetString(config.KeyLicensePlate)); err != nil {
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

	mailer := mail.NewService()

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
	reportCmd.Flags().StringVar(&monthFlag, "month", "", "Optional. Month in MM-YYYY format (e.g. 02-2026). Defaults to previous month.")
	reportCmd.Flags().StringVar(&fromMonthFlag, "from-month", "", "Start month in MM-YYYY format (e.g. 01-2026). Use with --to-month for multi-month reports.")
	reportCmd.Flags().StringVar(&toMonthFlag, "to-month", "", "End month in MM-YYYY format (e.g. 03-2026). Use with --from-month for multi-month reports.")
	reportCmd.Flags().BoolVar(&pdfFlag, "pdf", false, "Export the report as a PDF file.")
	reportCmd.Flags().BoolVar(&attachPdfsFlag, "attach-pdfs", false, fmt.Sprintf("Attach all PDF files from ~/%s/ to the generated report PDF. Requires --pdf.", config.ConfigDirName))
	reportCmd.Flags().BoolVar(&sendMailFlag, "send-mail", false, "Send the generated PDF via email. Requires --pdf and configured mail settings (-h for details).")

	rootCmd.AddCommand(reportCmd)
}

// getTimeRange parses the month flags and returns the start and end of the period along with a label.
// It returns an error if the flags are invalid.
func getTimeRange(monthFlag, fromMonthFlag, toMonthFlag string) (startOfPeriod, endOfPeriod time.Time, periodLabel string, err error) {
	if monthFlag != "" {
		// Single month mode (backward compatible)
		targetDate, parseErr := time.Parse("01-2006", monthFlag)
		if parseErr != nil {
			return time.Time{}, time.Time{}, "", fmt.Errorf("invalid date format for --month. Please use MM-YYYY (e.g. 02-2026)")
		}
		startOfPeriod = time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		endOfPeriod = startOfPeriod.AddDate(0, 1, 0).Add(-time.Nanosecond)
		periodLabel = monthFlag
	} else {
		// Multi-month mode
		fromDate, parseErr := time.Parse("01-2006", fromMonthFlag)
		if parseErr != nil {
			return time.Time{}, time.Time{}, "", fmt.Errorf("invalid date format for --from-month. Please use MM-YYYY (e.g. 02-2026)")
		}
		toDate, parseErr := time.Parse("01-2006", toMonthFlag)
		if parseErr != nil {
			return time.Time{}, time.Time{}, "", fmt.Errorf("invalid date format for --to-month. Please use MM-YYYY (e.g. 02-2026)")
		}

		if toDate.Before(fromDate) {
			return time.Time{}, time.Time{}, "", fmt.Errorf("--to-month must be equal to or after --from-month")
		}

		startOfPeriod = time.Date(fromDate.Year(), fromDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		endOfMonth := time.Date(toDate.Year(), toDate.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0).Add(-time.Nanosecond)
		endOfPeriod = endOfMonth
		periodLabel = fmt.Sprintf("%s_to_%s", fromMonthFlag, toMonthFlag)
	}
	return startOfPeriod, endOfPeriod, periodLabel, nil
}

func getPreviousMonth() string {
	now := time.Now()
	// Use the 1st of the current month to avoid day overflow when subtracting a month
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	prevMonth := firstOfMonth.AddDate(0, -1, 0)
	return prevMonth.Format("01-2006")
}
