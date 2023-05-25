package cmd

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/program"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Args:  cobra.MaximumNArgs(1),
	Short: "Initializes a new Keel project",
	Run: func(cmd *cobra.Command, args []string) {
		model := &program.InitModel{
			ProjectDir: args[0],
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
	rootCmd.AddCommand(initCmd)
}
