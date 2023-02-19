package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/node"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate project specific code",
	Run: func(cmd *cobra.Command, args []string) {
		files, err := node.Generate(context.Background(), flagProjectDir, node.WithDevelopmentServer(true))
		if err != nil {
			panic(err)
		}

		err = files.Write()
		if err != nil {
			panic(err)
		}

		fmt.Println("Done 🚀")
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
