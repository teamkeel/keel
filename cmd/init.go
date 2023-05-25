package cmd

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/program"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Args:  cobra.RangeArgs(0, 1),
	Short: "Initializes a new Keel project",
	Run: func(cmd *cobra.Command, args []string) {
		cwd, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		dir := cwd

		if len(args) > 0 {
			dir = args[0]
		}

		model := &program.InitModel{
			ProjectDir: dir,
		}

		_, err = tea.NewProgram(model).Run()
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
