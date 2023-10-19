package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/runtime"
)

var enabledDebugFlags = "true"

var (
	flagProjectDir       string
	flagReset            bool
	flagPort             string
	flagNodePackagesPath string
	flagPrivateKeyPath   string
	flagPattern          string
	flagTracing          bool
	flagVersion          bool
)

var rootCmd = &cobra.Command{
	Use:   "keel",
	Short: "The Keel CLI",
	RunE: func(cmd *cobra.Command, args []string) error {
		if flagVersion {
			fmt.Printf("v%s\n", runtime.GetVersion())
			os.Exit(0)
		}
		return cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	rootCmd.PersistentFlags().StringVarP(&flagProjectDir, "dir", "d", workingDir, "directory containing a Keel project")
	rootCmd.PersistentFlags().BoolVarP(&flagVersion, "version", "v", false, "Print the Keel CLI version")
}
