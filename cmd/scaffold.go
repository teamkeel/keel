package cmd

import (
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/program"
)

var scaffoldCmd = &cobra.Command{
	Use:   "scaffold",
	Short: "Scaffolds missing custom functions",
	Run: func(cmd *cobra.Command, args []string) {
		program.Run(&program.Model{
			Mode:       program.ModeScaffold,
			ProjectDir: flagProjectDir,
		})
	},
}

func init() {
	rootCmd.AddCommand(scaffoldCmd)
}
