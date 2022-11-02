package cmd

import (
	"fmt"
	"strings"

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

		if err != nil {
			return err
		}

		fmt.Print(workingDir)

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

		ch, err := testing.Run(workingDir, pattern)

		if err != nil {
			return err
		}

		results := []*testing.TestResult{}
		for newEvents := range ch {
			resultEvents := lo.Filter(newEvents, func(e *testing.Event, _ int) bool {
				return e.EventStatus == testing.EventStatusComplete && e.Result != nil
			})

			for _, e := range resultEvents {
				results = append(results, e.Result)
			}
		}

		PrintSummary(results)
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

func PrintSummary(results []*testing.TestResult) {
	totalPassed := lo.CountBy(results, func(evt *testing.TestResult) bool {
		return evt.Status == testing.StatusPass
	})

	totalFailed := lo.CountBy(results, func(evt *testing.TestResult) bool {
		return evt.Status != testing.StatusPass
	})

	for _, event := range results {
		if event.Status == testing.StatusPass {
			fmt.Printf("%s %s\n", color.New(color.BgGreen).Add(color.FgWhite).Sprint(" PASS "), event.TestName)
		} else {
			fmt.Printf("%s %s\n", color.New(color.BgRed).Add(color.FgWhite).Sprintf(" %s ", strings.ToUpper(event.Status)), event.TestName)

			switch event.Status {
			case testing.StatusFail:
				fmt.Printf("------------\n%s\n%s\n------------\n", color.New(color.FgGreen).Sprintf("Expected:\n%s", event.Expected), color.New(color.FgRed).Sprintf("Actual:\n%s", event.Actual))
			case testing.StatusException:
				fmt.Printf("------------\n%s\n------------\n", color.New(color.FgGreen).Sprintf("Error:\n%s", event.Err))
			}
		}
	}

	fmt.Printf("Test summary: %s, %s\n", color.New(color.FgGreen).Sprintf("%d passed", totalPassed), color.New(color.FgRed).Sprintf("%d failed", totalFailed))
}
