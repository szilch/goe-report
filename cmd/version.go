package cmd

import (
	"echarge-report/pkg/version"
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of echarge-report",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("echarge-report %s\n", version.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
