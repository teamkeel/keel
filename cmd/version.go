package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version string

// diffCmd represents the diff command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the Keel CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("v%s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
