package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/program"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate your project",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("unexpected arguments: %v", args)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		program.Run(&program.Model{
			Mode:       program.ModeValidate,
			ProjectDir: flagProjectDir,
		})
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
