package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/colors"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/schema"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates supporting SDK for a Keel schema and scaffolds missing custom functions",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		logPrefix := colors.Green("|").String()

		packageManager, err := resolvePackageManager(flagProjectDir, false)
		if err == promptui.ErrAbort {
			return nil
		}
		if err != nil {
			panic(err)
		}

		err = node.Bootstrap(
			flagProjectDir,
			node.WithPackageManager(packageManager),
			node.WithPackagesPath(flagNodePackagesPath),
			node.WithLogger(func(s string) {
				fmt.Println(logPrefix, s)
			}),
			node.WithOutputWriter(os.Stdout))
		if err != nil {
			return err
		}

		b := schema.Builder{}
		schema, err := b.MakeFromDirectory(flagProjectDir)
		if err != nil {
			return err
		}

		files, err := node.Generate(context.Background(), schema, node.WithDevelopmentServer(true))
		if err != nil {
			return err
		}

		err = files.Write(flagProjectDir)
		if err != nil {
			return err
		}

		fmt.Println(logPrefix, "Generated @teamkeel/sdk")
		fmt.Println(logPrefix, "Generated @teamkeel/testing")

		files, err = node.Scaffold(flagProjectDir, schema)
		if err != nil {
			return err
		}
		err = files.Write(flagProjectDir)
		if err != nil {
			return err
		}
		if len(files) > 0 {
			fmt.Println(logPrefix, "Scaffolded missing functions:")
			for _, f := range files {
				name := strings.TrimPrefix(f.Path, flagProjectDir)
				fmt.Println("  -", colors.Gray(name).String())
			}
		}

		fmt.Println(logPrefix, "Done ✨")
		fmt.Println("")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	if enabledDebugFlags == "true" {
		generateCmd.Flags().StringVar(&flagNodePackagesPath, "node-packages-path", "", "path to local @teamkeel npm packages")
	}
}
