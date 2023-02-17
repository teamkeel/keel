package cmd

import (
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/program"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate your project",
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
