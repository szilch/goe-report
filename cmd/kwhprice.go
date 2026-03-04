package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var kwhPriceCmd = &cobra.Command{
	Use:   "kwhprice",
	Short: "Manage the price per kWh",
	Long:  `Manage the price per kWh (in Euro) used to calculate reporting costs.`,
}

var kwhPriceSetCmd = &cobra.Command{
	Use:   "set [price]",
	Short: "Set the price per kWh (e.g. 0.35)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		priceStr := args[0]
		price, err := strconv.ParseFloat(priceStr, 64)
		if err != nil {
			fmt.Println("Error: Invalid price format. Please use a valid number (e.g. 0.35).")
			os.Exit(1)
		}

		viper.Set("kwhprice", price)

		err = viper.WriteConfig()
		if err != nil {
			err = viper.SafeWriteConfig()
		}

		if err != nil {
			fmt.Println("Error saving kWh price:", err)
			os.Exit(1)
		}
		fmt.Printf("kWh price (%.2f €) saved successfully.\n", price)
	},
}

var kwhPriceGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the current price per kWh",
	Run: func(cmd *cobra.Command, args []string) {
		price := viper.GetFloat64("kwhprice")
		if price == 0 {
			fmt.Println("No kWh price set (0.00 €).")
		} else {
			fmt.Printf("%.2f\n", price)
		}
	},
}

func init() {
	rootCmd.AddCommand(kwhPriceCmd)
	kwhPriceCmd.AddCommand(kwhPriceSetCmd)
	kwhPriceCmd.AddCommand(kwhPriceGetCmd)
}
