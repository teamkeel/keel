package cmd

import (
	"fmt"

	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/teamkeel/keel/nodedeps"
	"github.com/teamkeel/keel/testhelpers"
	"github.com/teamkeel/keel/testing"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run Keel tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		var ch chan []*testing.Event
		workingDir, err := testhelpers.WithTmpDir(inputDir)

		if err != nil {
			return err
		}

		onQuit := func() {
			if ch != nil {
				close(ch)
			}
		}
		outputter := testing.NewOutputter(workingDir, onQuit)

		outputter.Start()
		packageJson, err := nodedeps.NewPackageJson(filepath.Join(workingDir, "package.json"))

		if err != nil {
			return err
		}

		err = packageJson.Inject(map[string]string{
			"@teamkeel/testing": "*",
			"@teamkeel/sdk":     "*",
			"@teamkeel/runtime": "*",
			"ts-node":           "*",
			// https://typestrong.org/ts-node/docs/swc/
			"@swc/core":           "*",
			"regenerator-runtime": "*",
		}, true)

		if err != nil {
			return err
		}

		ch, err = testing.Run(workingDir, pattern)

		if err != nil {
			return err
		}

		for newEvents := range ch {
			outputter.Push(newEvents)
		}

		outputter.End()

		return nil
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
