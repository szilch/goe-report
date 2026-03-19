package cmd

import (
	"echarge-report/pkg/config"
	"fmt"
	"os"
	"path/filepath"

	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "echarge-report",
	Short: "Wallbox Charging Report CLI",
	Long: `echarge-report is a CLI tool for interacting with wallbox chargers.
It allows fetching status information as well as generating and 
exporting historical charging reports. Supports multiple wallbox types.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	defaultConfigPath := fmt.Sprintf("$HOME/%s/%s", config.ConfigDirName, config.ConfigFileName)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is %s)", defaultConfigPath))
}

func initConfig() {
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("ECHARGEREPORT")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		configDir := filepath.Join(home, config.ConfigDirName)
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			os.MkdirAll(configDir, 0755)
		}

		viper.SetConfigFile(filepath.Join(configDir, config.ConfigFileName))
	}
	if err := viper.ReadInConfig(); err != nil {
		if !os.IsNotExist(err) {
			color.Red("Error reading config file: %v", err)
		}
	}
}
