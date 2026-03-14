package cmd

import (
	"echarge-report/pkg/config"
	"echarge-report/pkg/formatter"
	"echarge-report/pkg/homeassistant"
	"echarge-report/pkg/mail"
	"echarge-report/pkg/pdf"
	"echarge-report/pkg/report"
	"echarge-report/pkg/wallbox"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var chipIdsFlag string
var monthFlag string
var fromMonthFlag string
var toMonthFlag string
var pdfFlag bool
var attachPdfsFlag bool
var sendMailFlag bool

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a charging report for a specific RFID and month or date range",
	Long:  `Fetches the charging history from the configured wallbox API and filters it by the provided RFID (or RFID Group) and month (in MM-YYYY format) or a date range (using --from-month and --to-month).`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create wallbox adapter using the factory
		adapter, err := wallbox.NewAdapter()
		if err != nil {
			color.Red("Error: %v", err)
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

		color.Blue("Fetching charging history for wallbox (type: %s)...", adapter.GetType())

		haService := homeassistant.NewService()
		reportSvc := report.NewService(adapter, haService)

		reportData, err := reportSvc.GenerateReportData(monthFlag, fromMonthFlag, toMonthFlag)
		if err != nil {
			color.Red("Error generating report data: %v", err)
			os.Exit(1)
		}

		var frm formatter.Formatter
		var reportFilename string
		if pdfFlag {
			reportFilename = fmt.Sprintf("echarge_report_%s.pdf", reportData.PeriodLabel)
			frm = formatter.NewPDFFormatter(reportFilename)
		} else {
			frm = formatter.NewTerminalFormatter()
		}

		if err := frm.Format(reportData); err != nil {
			color.Red("Error generating report output: %v", err)
			os.Exit(1)
		}

		// Attach PDFs from ~/.echarge-report/ if requested
		if pdfFlag && attachPdfsFlag {
			pdfSvc := pdf.NewService()
			attachedCount, configDir, err := pdfSvc.AttachExistingPDFsToReport(reportFilename)
			if err != nil {
				color.Red("%v", err)
				os.Exit(1)
			}
			if attachedCount == 0 {
				color.Yellow("Warning: --attach-pdfs set but no PDF files found in %s.", configDir)
			} else {
				color.Blue("Attaching PDFs...")
				color.Green("Attached %d PDF(s) from %s successfully.", attachedCount, configDir)
			}
		}

		// Send email if requested
		if sendMailFlag {
			color.Blue("Preparing to send email...")
			mailer := mail.NewService()
			if err := mailer.SendReportEmail(reportFilename, reportData); err != nil {
				color.Red("%v", err)
				os.Exit(1)
			}
			color.Green("Email sent successfully.")
		}
	},
}

func init() {
	reportCmd.Flags().StringVar(&chipIdsFlag, "chipIds", "", "Optional. Comma-separated list of chip IDs to filter by (e.g. 12345,67890)")
	reportCmd.Flags().StringVar(&monthFlag, "month", "", "Optional. Month in MM-YYYY format (e.g. 02-2026). Defaults to previous month.")
	reportCmd.Flags().StringVar(&fromMonthFlag, "from-month", "", "Start month in MM-YYYY format (e.g. 01-2026). Use with --to-month for multi-month reports.")
	reportCmd.Flags().StringVar(&toMonthFlag, "to-month", "", "End month in MM-YYYY format (e.g. 03-2026). Use with --from-month for multi-month reports.")
	reportCmd.Flags().BoolVar(&pdfFlag, "pdf", false, "Export the report as a PDF file.")
	reportCmd.Flags().BoolVar(&attachPdfsFlag, "attach-pdfs", false, fmt.Sprintf("Attach all PDF files from ~/%s/ to the generated report PDF. Requires --pdf.", config.ConfigDirName))
	reportCmd.Flags().BoolVar(&sendMailFlag, "send-mail", false, "Send the generated PDF via email. Requires --pdf and configured mail settings (-h for details).")

	viper.BindPFlag("month", reportCmd.Flags().Lookup("month"))
	viper.BindPFlag("from-month", reportCmd.Flags().Lookup("from-month"))
	viper.BindPFlag("to-month", reportCmd.Flags().Lookup("to-month"))

	rootCmd.AddCommand(reportCmd)
}
