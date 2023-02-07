package cmd

import (
	"fmt"

	"os"

	"github.com/spf13/cobra"

	"github.com/teamkeel/keel/cmd/database"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/testing"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run Keel tests",
	Run: func(cmd *cobra.Command, args []string) {

		opts := []node.BootstrapOption{}
		if os.Getenv("KEEL_LOCAL_PACKAGES_PATH") != "" {
			opts = append(opts, node.WithPackagesPath(os.Getenv("KEEL_LOCAL_PACKAGES_PATH")))
		}

		err := node.Bootstrap(inputDir, opts...)
		if err != nil {
			panic(err)
		}

		_, dbConnInfo, err := database.Start(true)
		if err != nil {
			panic(err)
		}
		defer func() {
			err = database.Stop()
			if err != nil {
				panic(err)
			}
		}()

		results, err := testing.Run(&testing.RunnerOpts{
			Dir:        inputDir,
			Pattern:    pattern,
			DbConnInfo: dbConnInfo,
		})

		if results != nil {
			fmt.Println(results.Output)
		}

		if err != nil {
			panic(err)
		}

		if results != nil && !results.Success {
			os.Exit(1)
		}
	},
}

var pattern string

func init() {
	rootCmd.AddCommand(testCmd)
	defaultDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("os.Getwd() errored: %v", err))
	}
	testCmd.Flags().StringVarP(&inputDir, "dir", "d", defaultDir, "input directory to validate")
	testCmd.Flags().StringVarP(&pattern, "pattern", "p", "(.*)", "pattern to isolate test")
}
