package cmd

import (
	"fmt"

	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/samber/lo"

	"github.com/spf13/cobra"

	"github.com/teamkeel/keel/nodedeps"
	"github.com/teamkeel/keel/testhelpers"
	"github.com/teamkeel/keel/testing"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run Keel tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		workingDir, err := testhelpers.WithTmpDir(inputDir)

		allEvents := []*testing.Event{}

		if err != nil {
			return err
		}

		packageJson, err := nodedeps.NewPackageJson(filepath.Join(workingDir, "package.json"))

		if err != nil {
			return err
		}

		err = packageJson.Inject(map[string]string{
			"@teamkeel/testing": "0.175.0",
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

		ch, err := testing.Run(workingDir, "", testing.RunTypeTestCmd)

		if err != nil {
			return err
		}

		// each js test file reports back an array of testing results
		for newEvents := range ch {
			allEvents = append(allEvents, newEvents...)
		}

		PrintSummary(allEvents)
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

func PrintSummary(events []*testing.Event) {

	totalPassed := lo.CountBy(events, func(evt *testing.Event) bool {
		return evt.Status == testing.StatusPass
	})

	totalFailed := lo.CountBy(events, func(evt *testing.Event) bool {
		return evt.Status != testing.StatusPass
	})

	fmt.Printf("Test summary: %s, %s\n", color.New(color.FgGreen).Sprintf("%d passed", totalPassed), color.New(color.FgRed).Sprintf("%d failed", totalFailed))

	// for _, event := range events {
	// 	if event.Status == testing.StatusPass {
	// 		fmt.Printf("%s %s\n", color.New(color.BgGreen).Add(color.FgWhite).Sprint(" PASS "), event.TestName)
	// 	} else {
	// 		fmt.Printf("%s %s\n", color.New(color.BgRed).Add(color.FgWhite).Sprintf(" %s ", event.Status), event.TestName)
	// 	}
	// }
}
