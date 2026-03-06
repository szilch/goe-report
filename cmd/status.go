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
