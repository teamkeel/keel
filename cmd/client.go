package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/program"
)

var flagClientPackage bool
var flagClientWatch bool
var flagClientOutputDir string
var flagClientApiName string

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Generates client SDK for a Keel project",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		model := &program.GenerateClientModel{
			ProjectDir: flagProjectDir,
			Package:    flagClientPackage,
			OutputDir:  flagClientOutputDir,
			ApiName:    flagClientApiName,
			Watch:      flagClientWatch,
		}

		_, err := tea.NewProgram(model).Run()
		if err != nil {
			return err
		}

		return model.Err
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)

	clientCmd.Flags().StringVarP(&flagClientApiName, "api", "a", "", "name of the API to generate a client for")
	clientCmd.Flags().StringVarP(&flagClientOutputDir, "output", "o", ".", "directory to output the client")
	clientCmd.Flags().BoolVar(&flagClientPackage, "package", false, "Set to true will generate a a client package, false will generate a single file client")
	clientCmd.Flags().BoolVar(&flagClientWatch, "watch", false, "Watch for schema changes and regenerate the client")
}
