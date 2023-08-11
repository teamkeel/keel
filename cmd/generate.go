package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/program"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates supporting SDK for a Keel schema and scaffolds missing custom functions",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return fmt.Errorf("unexpected arguments: %v", args)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		model := &program.GenerateModel{
			ProjectDir: flagProjectDir,
		}

		_, err := tea.NewProgram(model).Run()
		if err != nil {
			panic(err)
		}

		if model.Err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
