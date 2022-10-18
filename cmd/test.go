package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/teamkeel/keel/nodedeps"
	"github.com/teamkeel/keel/testhelpers"
	"github.com/teamkeel/keel/testing"
	testpackage "github.com/teamkeel/keel/testing"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run Keel tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		workingDir, err := testhelpers.WithTmpDir(inputDir)

		if err != nil {
			return err
		}

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

		ch, err := testpackage.Run(workingDir, "", testing.RunTypeTestCmd)

		if err != nil {
			return err
		}

		events := []*testing.Event{}
		for newEvents := range ch {
			events = append(events, newEvents...)

			fmt.Print(newEvents)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)
	defaultDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("os.Getwd() errored: %v", err))
	}
	testCmd.Flags().StringVarP(&inputDir, "dir", "d", defaultDir, "input directory to validate")
}
