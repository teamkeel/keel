package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/program"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a new Keel project",
	Run: func(cmd *cobra.Command, args []string) {
		dir, _ := os.Getwd()

		if len(args) > 0 {
			dir = args[0]
		}

		program.Run(&program.Model{
			Mode:       program.ModeInit,
			ProjectDir: dir,
		})
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
