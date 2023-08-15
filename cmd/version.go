package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/runtime"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the Keel CLI version",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("unexpected arguments: %v", args)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("v%s\n", runtime.GetVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
