package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/nodedeps"
	"github.com/teamkeel/keel/schema"
)

var NODE_MODULE_DIR string = ".keel"

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates code",
	Run: func(cmd *cobra.Command, args []string) {
		schemaDir, _ := cmd.Flags().GetString("dir")

		packageJson, err := nodedeps.NewPackageJson(filepath.Join(schemaDir, "package.json"))

		if err != nil {
			fmt.Println("⛔️ Could not create package.json automatically")
			return
		}

		err = packageJson.Bootstrap()

		if err != nil {
			fmt.Println("⛔️ Could not bootstrap package.json")
			return
		}

		b := schema.Builder{}

		schema, err := b.MakeFromDirectory(schemaDir)

		if err != nil {
			fmt.Println("⛔️ Could not read schema file")
		}

		r, err := functions.NewRuntime(schema, schemaDir)

		if err != nil {
			fmt.Println("⛔️ Internal runtime error (a)")
			fmt.Print(err)
			return
		}

		// We need to scaffold out any custom functions first
		// prior to generating the rest of the client code, which
		// references the custom functions via imports.
		sr, err := r.Scaffold()

		if err != nil {
			fmt.Println("⛔️ Internal runtime error (c)")
			fmt.Print(err)
			return
		}

		err = r.Bootstrap()

		if err != nil {
			fmt.Println("⛔️ Could not generate @teamkeel/client")
			fmt.Println(err)

			return
		}

		fmt.Printf("Generated the following files:\n\n")

		fmt.Printf("--- %s ---\n", color.New(color.FgHiYellow).Sprint("Functions"))

		if len(sr.CreatedFunctions) == 0 && sr.FunctionsCount > 0 {
			fmt.Println("✅  No new functions to generate")
		} else if sr.FunctionsCount == 0 {
			fmt.Println("✅  No custom functions defined")
		} else {
			for _, f := range sr.CreatedFunctions {
				fileName := filepath.Base(f)
				fmt.Printf("⚡️ Generated %s %s\n", color.New(color.FgCyan).Sprint(fileName), color.New(color.Faint).Sprintf("[%s]", f))
			}
		}
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
