package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var tokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage the go-e Wallbox API token",
	Long:  `Manage the API token to authenticate against the go-e Wallbox Cloud API.`,
}

var setCmd = &cobra.Command{
	Use:   "set [token]",
	Short: "Set the API token",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		token := args[0]
		viper.Set("token", token)

		err := viper.WriteConfig()
		if err != nil {
			err = viper.SafeWriteConfig()
		}

		if err != nil {
			color.Red("Error saving token: %v", err)
			os.Exit(1)
		}
		color.Blue("Token saved successfully.")
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the current API token",
	Run: func(cmd *cobra.Command, args []string) {
		token := viper.GetString("token")
		if token == "" {
			fmt.Println("No token set.")
		} else {
			fmt.Println(token)
		}
	},
}

func init() {
	rootCmd.AddCommand(tokenCmd)
	tokenCmd.AddCommand(setCmd)
	tokenCmd.AddCommand(getCmd)
}
