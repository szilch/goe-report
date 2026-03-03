package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the current status of the go-e Wallbox from the Cloud API",
	Long:  `Fetches the current status metrics from the go-e Cloud API using the saved token and serial number.`,
	Run: func(cmd *cobra.Command, args []string) {
		token := viper.GetString("token")
		serial := viper.GetString("serial")

		if token == "" || serial == "" {
			fmt.Println("Fehler: Token und Seriennummer müssen gesetzt sein.")
			fmt.Println("Nutze 'goe-report token set <token>' und 'goe-report serial set <serial>'.")
			os.Exit(1)
		}

		// According to go-e API v2 Cloud specifics, the URL format is:
		// https://<serial>.api.v3.go-e.io/api/status?token=<token>
		url := fmt.Sprintf("https://%s.api.v3.go-e.io/api/status?token=%s", serial, token)

		fmt.Printf("Frage Status für Wallbox %s ab...\n", serial)

		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Fehler beim Verbindungsaufbau zur go-e Cloud API: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Fehler beim Lesen der Antwort: %v\n", err)
			os.Exit(1)
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Fehler: Die API antwortete mit Statuscode %d\n", resp.StatusCode)
			fmt.Printf("Antwort: %s\n", string(body))
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
			fmt.Printf("Fehler beim Verarbeiten der JSON-Antwort: %v\n", err)
			os.Exit(1)
		}

		// Interpretation the 'car' state
		carState := "Unbekannt"
		switch statusData.Car {
		case 1:
			carState = "Frei (Nicht verbunden)"
		case 2:
			carState = "Lädt"
		case 3:
			carState = "Wartet auf Auto"
		case 4:
			carState = "Laden beendet"
		case 5:
			carState = "Fehler"
		}

		// Allowed state
		alwState := "Nein"
		if statusData.Alw {
			alwState = "Ja"
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
		fmt.Println("\nWallbox Status Bericht:")
		fmt.Println("--------------------------------------------------")

		fmt.Printf("%-25s %s\n", "Status Fahrzeug:", carState)
		fmt.Printf("%-25s %s\n", "Laden erlaubt:", alwState)
		fmt.Printf("%-25s %d A\n", "Eingestellte Stromstärke:", statusData.Amp)
		fmt.Printf("%-25s %.2f kW\n", "Aktuelle Leistung:", pTotal/1000.0)
		fmt.Printf("%-25s %.2f kWh\n", "Geladen seit Anstecken:", statusData.Wh/1000.0)
		if statusData.Eto > 0 {
			fmt.Printf("%-25s %.2f kWh\n", "Gesamtverbrauch (Total):", statusData.Eto/1000.0)
		}
		fmt.Printf("%-25s %s\n", "Gerätetemperatur:", tempStr)

		// Print Phasen details if available
		if numNrg >= 10 {
			fmt.Println("\nDetails nach Phasen:")
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
