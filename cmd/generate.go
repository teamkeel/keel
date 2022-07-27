package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/functions"
)

var NODE_MODULE_DIR string = ".keel"

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates code",
	RunE: func(cmd *cobra.Command, args []string) error {
		schemaDir, _ := cmd.Flags().GetString("dir")

		r, err := functions.NewRuntime(schemaDir)

		if err != nil {
			return err
		}

		err = r.GenerateClient()

		if err != nil {
			return err
		}

		err = r.GenerateHandler()

		if err != nil {
			return err
		}

		result, errs := r.Bundle(true)

		if len(errs) > 0 {
			return err
		}

		fmt.Println("🔨 Generating code...")

		fmt.Println("---")

		for _, f := range result.OutputFiles {
			lastFragment := filepath.Base(f.Path)

			if err != nil {
				return err
			}
			fmt.Printf("⚡️ Generated %s [%s]\n", lastFragment, color.New(color.FgCyan).Sprint(f.Path))
		}

		fmt.Println("---")

		fmt.Println("✅ Generation complete")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	defaultDir, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	generateCmd.Flags().StringVarP(&inputDir, "dir", "d", defaultDir, "the directory containing the Keel schema files")
}
