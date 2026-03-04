package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var licensePlateCmd = &cobra.Command{
	Use:   "licenseplate",
	Short: "Manage the vehicle license plate",
	Long:  `Manage the vehicle license plate (Kennzeichen) to be used in reports.`,
}

var licensePlateSetCmd = &cobra.Command{
	Use:   "set [licenseplate]",
	Short: "Set the vehicle license plate",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		licensePlate := args[0]
		viper.Set("licenseplate", licensePlate)

		err := viper.WriteConfig()
		if err != nil {
			err = viper.SafeWriteConfig()
		}

		if err != nil {
			fmt.Println("Error saving license plate:", err)
			os.Exit(1)
		}
		fmt.Println("License plate saved successfully.")
	},
}

var licensePlateGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the current vehicle license plate",
	Run: func(cmd *cobra.Command, args []string) {
		licensePlate := viper.GetString("licenseplate")
		if licensePlate == "" {
			fmt.Println("No license plate set.")
		} else {
			fmt.Println(licensePlate)
		}
	},
}

func init() {
	rootCmd.AddCommand(licensePlateCmd)
	licensePlateCmd.AddCommand(licensePlateSetCmd)
	licensePlateCmd.AddCommand(licensePlateGetCmd)
}
