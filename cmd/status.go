package cmd

import (
	"echarge-report/pkg/config"
	"echarge-report/pkg/wallbox"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the current status of the configured Wallbox",
	Long:  `Fetches the current status metrics from the configured wallbox API.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Create wallbox adapter using the factory
		adapter, err := wallbox.NewAdapter()
		if err != nil {
			color.Red("Error: %v", err)
			os.Exit(1)
		}

		serial := viper.GetString(config.KeyWallboxSerial)

		// Validate configuration based on wallbox type
		if adapter.GetType() == "goe" {
			token := viper.GetString(config.KeyWallboxToken)
			localApiUrl := viper.GetString(config.KeyWallboxLocalApiUrl)

			if (token == "" || serial == "") && localApiUrl == "" {
				color.Red("Error: Either a Cloud API Token or a Local API URL must be configured.")
				color.Red("Use 'echarge-report config-set wallbox_token <token>' and 'echarge-report config-set wallbox_serial <serial>' or 'echarge-report config-set wallbox_localApiUrl http://<ip>'.")
				os.Exit(1)
			}
		}

		color.Blue("Fetching status for wallbox %s (type: %s)...", serial, adapter.GetType())

		statusData, err := adapter.GetStatus()
		if err != nil {
			color.Red("Failed to retrieve status: %v", err)
			os.Exit(1)
		}

		// Print formatted output
		fmt.Println("\nWallbox Status Report:")
		fmt.Println("--------------------------------------------------")

		fmt.Printf("%-25s %s\n", "Vehicle state:", statusData.VehicleState)
		fmt.Printf("%-25s %s\n", "Charging allowed:", statusData.ChargingAllowed)
		fmt.Printf("%-25s %d A\n", "Set current:", statusData.SetCurrentA)
		fmt.Printf("%-25s %.2f kW\n", "Current power:", statusData.CurrentPowerKW)
		fmt.Printf("%-25s %.2f kWh\n", "Charged since plug-in:", statusData.ChargedSincePlugInKWh)
		if statusData.TotalEnergyLifetimeKWh > 0 {
			fmt.Printf("%-25s %.2f kWh\n", "Total energy (lifetime):", statusData.TotalEnergyLifetimeKWh)
		}
		fmt.Printf("%-25s %s\n", "Device temperature:", statusData.TemperatureCelsius)

		// Print phase details if available
		if len(statusData.Phases) == 3 {
			fmt.Println("\nPhase details:")
			fmt.Printf("  L1: %5.1f V | %5.1f A | %5.0f W\n", statusData.Phases[0].Voltage, statusData.Phases[0].Current, statusData.Phases[0].Power)
			fmt.Printf("  L2: %5.1f V | %5.1f A | %5.0f W\n", statusData.Phases[1].Voltage, statusData.Phases[1].Current, statusData.Phases[1].Power)
			fmt.Printf("  L3: %5.1f V | %5.1f A | %5.0f W\n", statusData.Phases[2].Voltage, statusData.Phases[2].Current, statusData.Phases[2].Power)
		}
		fmt.Println("--------------------------------------------------")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
