package cmd

import (
	"fmt"
	"goe-report/pkg/config"
	"goe-report/pkg/goe"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the current status of the go-e Wallbox from the Cloud API",
	Long:  `Fetches the current status metrics from the go-e Cloud API using the saved token and serial number.`,
	Run: func(cmd *cobra.Command, args []string) {
		token := viper.GetString(config.KeyToken)
		serial := viper.GetString(config.KeySerial)
		localApiUrl := viper.GetString(config.KeyLocalApiUrl)

		if (token == "" || serial == "") && localApiUrl == "" {
			color.Red("Error: Either a Cloud API Token or a Local API URL must be configured.")
			color.Red("Use 'goe-report config-set goe_token <token>' and 'goe-report config-set goe_serial <serial>' or 'goe-report config-set goe_localApiUrl http://<ip>'.")
			os.Exit(1)
		}

		color.Blue("Fetching status for wallbox %s...", serial)

		client := goe.NewClient(serial, token, localApiUrl)
		statusData, err := client.GetStatus()
		if err != nil {
			color.Red("Failed to retrieve status: %v", err)
			os.Exit(1)
		}

		// Interpret the 'car' state
		carState := "Unknown"
		switch statusData.Car {
		case 1:
			carState = "Idle (not connected)"
		case 2:
			carState = "Charging"
		case 3:
			carState = "Waiting for car"
		case 4:
			carState = "Charging complete"
		case 5:
			carState = "Error"
		}

		// Allowed state
		alwState := "No"
		if statusData.Alw {
			alwState = "Yes"
		}

		// Calculate total power from NRG array
		var pTotal float64 = 0
		var numNrg = len(statusData.Nrg)
		if numNrg >= 12 {
			pTotal = statusData.Nrg[11] // in Watts
		} else if numNrg >= 4 {
			// fallback if array is smaller but has power (often index 7,8,9 are power)
			// normally index 11 is total power in v3/v4
		}

		// Temperature
		var tempStr = "N/A"
		if len(statusData.Tma) > 0 {
			tempStr = fmt.Sprintf("%.1f °C", statusData.Tma[0])
		}

		// Print formatted output
		fmt.Println("\nWallbox Status Report:")
		fmt.Println("--------------------------------------------------")

		fmt.Printf("%-25s %s\n", "Vehicle state:", carState)
		fmt.Printf("%-25s %s\n", "Charging allowed:", alwState)
		fmt.Printf("%-25s %d A\n", "Set current:", statusData.Amp)
		fmt.Printf("%-25s %.2f kW\n", "Current power:", pTotal/1000.0)
		fmt.Printf("%-25s %.2f kWh\n", "Charged since plug-in:", statusData.Wh/1000.0)
		if statusData.Eto > 0 {
			fmt.Printf("%-25s %.2f kWh\n", "Total energy (lifetime):", statusData.Eto/1000.0)
		}
		fmt.Printf("%-25s %s\n", "Device temperature:", tempStr)

		// Print phase details if available
		if numNrg >= 10 {
			fmt.Println("\nPhase details:")
			fmt.Printf("  L1: %5.1f V | %5.1f A | %5.0f W\n", statusData.Nrg[0], statusData.Nrg[4], statusData.Nrg[7])
			fmt.Printf("  L2: %5.1f V | %5.1f A | %5.0f W\n", statusData.Nrg[1], statusData.Nrg[5], statusData.Nrg[8])
			fmt.Printf("  L3: %5.1f V | %5.1f A | %5.0f W\n", statusData.Nrg[2], statusData.Nrg[6], statusData.Nrg[9])
		}
		fmt.Println("--------------------------------------------------")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
