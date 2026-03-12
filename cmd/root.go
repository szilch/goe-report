package cmd

import (
	"echarge-report/pkg/config"
	"fmt"
	"os"
	"path/filepath"

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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	defaultConfigPath := fmt.Sprintf("$HOME/%s/%s", config.ConfigDirName, config.ConfigFileName)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("config file (default is %s)", defaultConfigPath))
}

func initConfig() {
	viper.SetConfigType("env")
	viper.SetEnvPrefix("ECHARGEREPORT")
	viper.AutomaticEnv()
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in ~/.echarge-report/.echargereportrc
		configDir := filepath.Join(home, config.ConfigDirName)
		if _, err := os.Stat(configDir); os.IsNotExist(err) {
			os.MkdirAll(configDir, 0755)
		}

		// Set Viper config
		viper.SetConfigFile(filepath.Join(configDir, config.ConfigFileName))
	}
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
