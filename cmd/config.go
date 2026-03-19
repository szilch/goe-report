package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"echarge-report/pkg/config"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type configKey struct {
	Key         string
	Description string
}

var allowedKeys = []configKey{
	{Key: config.KeyWallboxGoeCloudToken, Description: "go-e Cloud API token"},
	{Key: config.KeyWallboxGoeLocalApiUrl, Description: "go-e Local API URL (e.g. http://192.168.1.50)"},
	{Key: config.KeyWallboxGoeCloudSerial, Description: "go-e Wallbox serial number"},
	{Key: config.KeyWallboxGoeChipIds, Description: "Chip IDs to filter by"},
	{Key: config.KeyLicensePlate, Description: "License plate (shown in the report)"},
	{Key: config.KeyDriver, Description: "Name of the driver (shown in the report)"},
	{Key: config.KeyKwhPrice, Description: "Price per kWh in EUR (e.g. 0.35)"},
	{Key: config.KeyHAToken, Description: "Home Assistant long-lived access token"},
	{Key: config.KeyHAWsHost, Description: "Home Assistant WebSocket Host (e.g. ws://homeassistant.local:8123)"},
	{Key: config.KeyHAMilageSensor, Description: "Home Assistant entity ID of the mileage sensor"},
	{Key: config.KeyMailHost, Description: "SMTP Host (e.g. smtp.example.com)"},
	{Key: config.KeyMailPort, Description: "SMTP Port (e.g. 587)"},
	{Key: config.KeyMailUsername, Description: "SMTP Username (e.g. user@example.com)"},
	{Key: config.KeyMailPassword, Description: "SMTP Password"},
	{Key: config.KeyMailFrom, Description: "Sender email address (e.g. sender@example.com)"},
	{Key: config.KeyMailTo, Description: "Comma-separated list of recipient email addresses"},
}

func isAllowedKey(key string) (*configKey, bool) {
	for i, k := range allowedKeys {
		if strings.EqualFold(k.Key, key) {
			return &allowedKeys[i], true
		}
	}
	return nil, false
}

func keyList() string {
	var sb strings.Builder
	for _, k := range allowedKeys {
		fmt.Fprintf(&sb, "  %-40s %s\n", k.Key, k.Description)
	}
	return sb.String()
}

var configSetCmd = &cobra.Command{
	Use:   "config-set <key> <value>",
	Short: "Set a configuration value",
	Long:  fmt.Sprintf("Sets a configuration value and persists it permanently.\n\nAllowed keys:\n%s", keyList()),
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := strings.ToLower(args[0])
		value := args[1]

		if _, ok := isAllowedKey(key); !ok {
			color.New(color.FgRed).Fprintf(cmd.OutOrStdout(), "Error: Unknown key \"%s\".\n", key)
			color.New(color.FgRed).Fprintf(cmd.OutOrStdout(), "Allowed keys:\n%s\n", keyList())
			os.Exit(1)
		}

		viper.Set(key, value)

		home, _ := os.UserHomeDir()
		configDir := filepath.Join(home, config.ConfigDirName)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			color.New(color.FgRed).Fprintf(cmd.OutOrStdout(), "Error creating config directory: %v\n", err)
			os.Exit(1)
		}

		if err := viper.WriteConfig(); err != nil {
			if err := viper.SafeWriteConfig(); err != nil {
				color.New(color.FgRed).Fprintf(cmd.OutOrStdout(), "Error saving configuration: %v\n", err)
				os.Exit(1)
			}
		}

		color.New(color.FgBlue).Fprintf(cmd.OutOrStdout(), "Configuration saved: %s = %s\n", key, value)
	},
}

var configGetCmd = &cobra.Command{
	Use:   "config-get <key>",
	Short: "Read a configuration value",
	Long:  fmt.Sprintf("Reads a stored configuration value.\n\nAllowed keys:\n%s", keyList()),
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := strings.ToLower(args[0])

		if _, ok := isAllowedKey(key); !ok {
			color.New(color.FgRed).Fprintf(cmd.OutOrStdout(), "Error: Unknown key \"%s\".\n", key)
			color.New(color.FgRed).Fprintf(cmd.OutOrStdout(), "Allowed keys:\n%s\n", keyList())
			os.Exit(1)
		}

		value := viper.GetString(key)
		if value == "" {
			cmd.Printf("(not set)\n")
		} else {
			cmd.Printf("%s\n", value)
		}
	},
}

var configListCmd = &cobra.Command{
	Use:   "config-list",
	Short: "Show all configuration values",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Println("Current configuration:")
		cmd.Println(strings.Repeat("-", 65))
		for _, k := range allowedKeys {
			val := viper.GetString(k.Key)
			if val == "" {
				val = "(not set)"
			}
			cmd.Printf("  %-40s %s\n", k.Key, val)
		}
		cmd.Println(strings.Repeat("-", 65))
	},
}

func init() {
	rootCmd.AddCommand(configSetCmd)
	rootCmd.AddCommand(configGetCmd)
	rootCmd.AddCommand(configListCmd)
}
