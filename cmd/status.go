package cmd

import (
	"echarge-report/pkg/config"
	"echarge-report/pkg/wallbox"
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
		adapter, err := wallbox.NewAdapter()
		if err != nil {
			color.New(color.FgRed).Fprintf(cmd.OutOrStdout(), "Error: %v\n", err)
			os.Exit(1)
		}

		serial := viper.GetString(config.KeyWallboxGoeCloudSerial)

		if adapter.GetType() == "goe" {
			token := viper.GetString(config.KeyWallboxGoeCloudToken)
			localApiUrl := viper.GetString(config.KeyWallboxGoeLocalApiUrl)

			if (token == "" || serial == "") && localApiUrl == "" {
				color.New(color.FgRed).Fprintf(cmd.OutOrStdout(), "Error: Either a Cloud API Token or a Local API URL must be configured.\n")
				color.New(color.FgRed).Fprintf(cmd.OutOrStdout(), "Use 'echarge-report config-set wallbox.goe.cloud.token <token>' and 'echarge-report config-set wallbox.goe.cloud.serial <serial>' or 'echarge-report config-set wallbox.goe.local.apiUrl http://<ip>'.\n")
				os.Exit(1)
			}
		}

		color.New(color.FgBlue).Fprintf(cmd.OutOrStdout(), "Fetching status for wallbox %s (type: %s)...\n", serial, adapter.GetType())

		statusData, err := adapter.GetStatus()
		if err != nil {
			color.New(color.FgRed).Fprintf(cmd.OutOrStdout(), "Failed to retrieve status: %v\n", err)
			os.Exit(1)
		}

		cmd.Println("\nWallbox Status Report:")
		cmd.Println("--------------------------------------------------")

		cmd.Printf("%-25s %s\n", "Vehicle state:", statusData.VehicleState)
		cmd.Printf("%-25s %t\n", "Charging allowed:", statusData.ChargingAllowed)
		cmd.Printf("%-25s %d A\n", "Set current:", statusData.SetCurrentA)
		cmd.Printf("%-25s %.2f kW\n", "Current power:", statusData.CurrentPowerKW)
		cmd.Printf("%-25s %.2f kWh\n", "Charged since plug-in:", statusData.ChargedSincePlugInKWh)
		if statusData.TotalEnergyLifetimeKWh > 0 {
			cmd.Printf("%-25s %.2f kWh\n", "Total energy (lifetime):", statusData.TotalEnergyLifetimeKWh)
		}
		cmd.Printf("%-25s %s\n", "Device temperature:", statusData.TemperatureCelsius)

		if len(statusData.Phases) == 3 {
			cmd.Println("\nPhase details:")
			cmd.Printf("  L1: %5.1f V | %5.1f A | %5.0f W\n", statusData.Phases[0].Voltage, statusData.Phases[0].Current, statusData.Phases[0].Power)
			cmd.Printf("  L2: %5.1f V | %5.1f A | %5.0f W\n", statusData.Phases[1].Voltage, statusData.Phases[1].Current, statusData.Phases[1].Power)
			cmd.Printf("  L3: %5.1f V | %5.1f A | %5.0f W\n", statusData.Phases[2].Voltage, statusData.Phases[2].Current, statusData.Phases[2].Power)
		}
		cmd.Println("--------------------------------------------------")
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
