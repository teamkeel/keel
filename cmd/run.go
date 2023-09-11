package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/program"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run your Keel App for development",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("unexpected arguments: %v", args)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		program.Run(&program.Model{
			Mode:             program.ModeRun,
			ProjectDir:       flagProjectDir,
			ResetDatabase:    flagReset,
			Port:             flagPort,
			TracingEnabled:   flagTracing,
			NodePackagesPath: flagNodePackagesPath,
			PrivateKeyPath:   flagPrivateKeyPath,
		})
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().BoolVar(&flagReset, "reset", false, "if set the database will be reset")
	runCmd.Flags().StringVar(&flagPort, "port", "8000", "the port to run the Keel application on")
	runCmd.Flags().StringVar(&flagPrivateKeyPath, "private-key-path", "", "path to the private key .pem file")

	if enabledDebugFlags == "true" {
		runCmd.Flags().StringVar(&flagNodePackagesPath, "node-packages-path", "", "path to local @teamkeel npm packages")
		runCmd.Flags().BoolVar(&flagTracing, "tracing", false, "enable tracing - an OTEL collector (e.g. jaeger) must be running on localhost:4318")
	}
}
