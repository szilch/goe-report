package cmd

import (
	"fmt"
	"os"
	"strings"

	"goe-report/pkg/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configKey beschreibt einen zulässigen Konfigurationsschlüssel.
type configKey struct {
	Key         string // Viper-Schlüssel
	Description string // Kurzbeschreibung für die Hilfe
}

// allowedKeys enthält alle Konfigurationsattribute, die gesetzt / gelesen werden dürfen.
var allowedKeys = []configKey{
	{Key: config.KeyToken, Description: "go-e Cloud API Token"},
	{Key: config.KeySerial, Description: "Seriennummer der Wallbox"},
	{Key: config.KeyLicensePlate, Description: "Kfz-Kennzeichen (wird im Report angezeigt)"},
	{Key: config.KeyKwhPrice, Description: "Preis pro kWh in Euro (z.B. 0.35)"},
	{Key: config.KeyHAToken, Description: "Home Assistant Long-Lived Access Token"},
	{Key: config.KeyHAAPI, Description: "Home Assistant API-URL (z.B. https://homeassistant.local:8123)"},
	{Key: config.KeyHAMilageSensor, Description: "Home Assistant Entity-ID des Kilometerstand-Sensors"},
}

// isAllowedKey prüft, ob ein Schlüssel erlaubt ist und gibt ihn ggf. zurück.
func isAllowedKey(key string) (*configKey, bool) {
	for i, k := range allowedKeys {
		if strings.EqualFold(k.Key, key) {
			return &allowedKeys[i], true
		}
	}
	return nil, false
}

// keyList gibt eine formatierte Übersicht aller erlaubten Schlüssel zurück.
func keyList() string {
	var sb strings.Builder
	for _, k := range allowedKeys {
		sb.WriteString(fmt.Sprintf("  %-25s %s\n", k.Key, k.Description))
	}
	return sb.String()
}

// --- set ---

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Konfigurationswert setzen",
	Long:  fmt.Sprintf("Setzt einen Konfigurationswert und speichert ihn dauerhaft.\n\nErlaubte Schlüssel:\n%s", keyList()),
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := strings.ToLower(args[0])
		value := args[1]

		if _, ok := isAllowedKey(key); !ok {
			color.Red("Fehler: Unbekannter Schlüssel \"%s\".", key)
			color.Red("Erlaubte Schlüssel:\n%s", keyList())
			os.Exit(1)
		}

		viper.Set(key, value)

		err := viper.WriteConfig()
		if err != nil {
			err = viper.SafeWriteConfig()
		}
		if err != nil {
			color.Red("Fehler beim Speichern der Konfiguration: %v", err)
			os.Exit(1)
		}

		color.Blue("Konfiguration gespeichert: %s = %s", key, value)
	},
}

// --- get ---

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Konfigurationswert lesen",
	Long:  fmt.Sprintf("Liest einen gespeicherten Konfigurationswert.\n\nErlaubte Schlüssel:\n%s", keyList()),
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := strings.ToLower(args[0])

		if _, ok := isAllowedKey(key); !ok {
			color.Red("Fehler: Unbekannter Schlüssel \"%s\".", key)
			color.Red("Erlaubte Schlüssel:\n%s", keyList())
			os.Exit(1)
		}

		value := viper.GetString(key)
		if value == "" {
			fmt.Printf("(nicht gesetzt)\n")
		} else {
			fmt.Println(value)
		}
	},
}

// --- list (Bonus: alle Werte auf einmal anzeigen) ---

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "Alle Konfigurationswerte anzeigen",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Aktuelle Konfiguration:")
		fmt.Println(strings.Repeat("-", 55))
		for _, k := range allowedKeys {
			val := viper.GetString(k.Key)
			if val == "" {
				val = "(nicht gesetzt)"
			}
			fmt.Printf("  %-25s %s\n", k.Key, val)
		}
		fmt.Println(strings.Repeat("-", 55))
	},
}

func init() {
	rootCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configGetCmd)
	rootCmd.AddCommand(configListCmd)
}
