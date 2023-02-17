package cmd

import (
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/program"
)

var Debug = true

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run your Keel App for development",
	Run: func(cmd *cobra.Command, args []string) {
		program.Run(&program.Model{
			Mode:             program.ModeRun,
			ProjectDir:       flagProjectDir,
			ResetDatabase:    flagReset,
			Port:             flagPort,
			NodePackagesPath: flagNodePackagesPath,
		})
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().BoolVar(&flagReset, "reset", false, "if set the database will be reset")
	runCmd.Flags().StringVar(&flagPort, "port", "8000", "the port to run the Keel application on")

	if Debug {
		runCmd.Flags().StringVar(&flagNodePackagesPath, "node-packages-path", "", "path to local @teamkeel npm packages")
	}
}
