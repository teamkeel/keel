package cmd

import (
	"fmt"

	"github.com/manifoldco/promptui"
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
		packageManager, err := resolvePackageManager(flagProjectDir, false)
		if err == promptui.ErrAbort {
			return
		}
		if err != nil {
			panic(err)
		}

		program.Run(&program.Model{
			Mode:             program.ModeRun,
			ProjectDir:       flagProjectDir,
			ResetDatabase:    flagReset,
			Port:             flagPort,
			CustomTracing:    flagTracing,
			NodePackagesPath: flagNodePackagesPath,
			PackageManager:   packageManager,
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
		runCmd.Flags().BoolVar(&flagTracing, "tracing", false, "trace instead with an OTEL collector (e.g. jaeger) on localhost:4318")
		runCmd.Flags().BoolVar(&flagVerboseTracing, "verbose-tracing", false, "display all events in tracing")
	}
}
