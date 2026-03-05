package cmd

import (
	"encoding/json"
	"fmt"
	"goe-report/pkg/config"
	"io"
	"net/http"
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

		if token == "" || serial == "" {
			color.Red("Error: Token and serial number must be set.")
			color.Red("Use 'goe-report config-set goe_token <token>' and 'goe-report config-set goe_serial <serial>'.")
			os.Exit(1)
		}

		// According to go-e API v2 Cloud specifics, the URL format is:
		// https://<serial>.api.v3.go-e.io/api/status?token=<token>
		url := fmt.Sprintf("https://%s.api.v3.go-e.io/api/status?token=%s", serial, token)

		color.Blue("Fetching status for wallbox %s...", serial)

		resp, err := http.Get(url)
		if err != nil {
			color.Red("Error connecting to the go-e Cloud API: %v", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			color.Red("Error reading response: %v", err)
			os.Exit(1)
		}

		if resp.StatusCode != http.StatusOK {
			color.Red("Error: The API responded with status code %d", resp.StatusCode)
			color.Red("Response: %s", string(body))
			os.Exit(1)
		}

		var statusData struct {
			Car int       `json:"car"` // 1: idle, 2: charging, 3: wait car, 4: complete, 5: error
			Alw bool      `json:"alw"`
			Amp int       `json:"amp"`
			Wh  float64   `json:"wh"`
			Eto float64   `json:"eto"`
			Nrg []float64 `json:"nrg"`
			Tma []float64 `json:"tma"`
			Frc int       `json:"frc"` // Force state
		}

		if err := json.Unmarshal(body, &statusData); err != nil {
			color.Red("Error processing JSON response: %v", err)
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
