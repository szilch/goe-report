package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serialCmd = &cobra.Command{
	Use:   "serial",
	Short: "Manage the go-e Wallbox serial number",
	Long:  `Manage the serial number of the go-e Wallbox.`,
}

var serialSetCmd = &cobra.Command{
	Use:   "set [serial]",
	Short: "Set the wallbox serial number",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serial := args[0]
		viper.Set("serial", serial)

		err := viper.WriteConfig()
		if err != nil {
			err = viper.SafeWriteConfig()
		}

		if err != nil {
			color.Red("Error saving serial number: %v", err)
			os.Exit(1)
		}
		color.Blue("Serial number saved successfully.")
	},
}

var serialGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the current wallbox serial number",
	Run: func(cmd *cobra.Command, args []string) {
		serial := viper.GetString("serial")
		if serial == "" {
			fmt.Println("No serial number set.")
		} else {
			fmt.Println(serial)
		}
	},
}

func init() {
	rootCmd.AddCommand(serialCmd)
	serialCmd.AddCommand(serialSetCmd)
	serialCmd.AddCommand(serialGetCmd)
}
