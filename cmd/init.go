package cmd

import (
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/program"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a new Keel project",
	Run: func(cmd *cobra.Command, args []string) {
		program.Run(&program.Model{
			Mode:       program.ModeInit,
			ProjectDir: flagProjectDir,
		})
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
