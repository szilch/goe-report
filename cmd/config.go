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

// configKey describes an allowed configuration key.
type configKey struct {
	Key         string // Viper key
	Description string // Short description for the help text
}

// allowedKeys contains all configuration attributes that may be set or read.
var allowedKeys = []configKey{
	{Key: config.KeyToken, Description: "go-e Cloud API token"},
	{Key: config.KeySerial, Description: "Wallbox serial number"},
	{Key: config.KeyChipIds, Description: "Default comma-separated list of chip IDs to filter by"},
	{Key: config.KeyLicensePlate, Description: "License plate (shown in the report)"},
	{Key: config.KeyKwhPrice, Description: "Price per kWh in EUR (e.g. 0.35)"},
	{Key: config.KeyHAToken, Description: "Home Assistant long-lived access token"},
	{Key: config.KeyHAAPI, Description: "Home Assistant API URL (e.g. https://homeassistant.local:8123)"},
	{Key: config.KeyHAMilageSensor, Description: "Home Assistant entity ID of the mileage sensor"},
	{Key: config.KeyMailHost, Description: "SMTP Host (e.g. smtp.example.com)"},
	{Key: config.KeyMailPort, Description: "SMTP Port (e.g. 587)"},
	{Key: config.KeyMailUsername, Description: "SMTP Username (e.g. user@example.com)"},
	{Key: config.KeyMailPassword, Description: "SMTP Password"},
	{Key: config.KeyMailFrom, Description: "Sender email address (e.g. sender@example.com)"},
	{Key: config.KeyMailTo, Description: "Comma-separated list of recipient email addresses"},
}

// isAllowedKey checks whether a key is allowed and returns it if so.
func isAllowedKey(key string) (*configKey, bool) {
	for i, k := range allowedKeys {
		if strings.EqualFold(k.Key, key) {
			return &allowedKeys[i], true
		}
	}
	return nil, false
}

// keyList returns a formatted overview of all allowed keys.
func keyList() string {
	var sb strings.Builder
	for _, k := range allowedKeys {
		sb.WriteString(fmt.Sprintf("  %-25s %s\n", k.Key, k.Description))
	}
	return sb.String()
}

// --- config-set ---

var configSetCmd = &cobra.Command{
	Use:   "config-set <key> <value>",
	Short: "Set a configuration value",
	Long:  fmt.Sprintf("Sets a configuration value and persists it permanently.\n\nAllowed keys:\n%s", keyList()),
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := strings.ToLower(args[0])
		value := args[1]

		if _, ok := isAllowedKey(key); !ok {
			color.Red("Error: Unknown key \"%s\".", key)
			color.Red("Allowed keys:\n%s", keyList())
			os.Exit(1)
		}

		viper.Set(key, value)

		err := viper.WriteConfig()
		if err != nil {
			err = viper.SafeWriteConfig()
		}
		if err != nil {
			color.Red("Error saving configuration: %v", err)
			os.Exit(1)
		}

		color.Blue("Configuration saved: %s = %s", key, value)
	},
}

// --- config-get ---

var configGetCmd = &cobra.Command{
	Use:   "config-get <key>",
	Short: "Read a configuration value",
	Long:  fmt.Sprintf("Reads a stored configuration value.\n\nAllowed keys:\n%s", keyList()),
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := strings.ToLower(args[0])

		if _, ok := isAllowedKey(key); !ok {
			color.Red("Error: Unknown key \"%s\".", key)
			color.Red("Allowed keys:\n%s", keyList())
			os.Exit(1)
		}

		value := viper.GetString(key)
		if value == "" {
			fmt.Printf("(not set)\n")
		} else {
			fmt.Println(value)
		}
	},
}

// --- config-list ---

var configListCmd = &cobra.Command{
	Use:   "config-list",
	Short: "Show all configuration values",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Current configuration:")
		fmt.Println(strings.Repeat("-", 55))
		for _, k := range allowedKeys {
			val := viper.GetString(k.Key)
			if val == "" {
				val = "(not set)"
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
