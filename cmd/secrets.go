package cmd

import (
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/program"
)

var (
	environment string
)

// secretsCmd represents the secrets command
var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Interact with your Keel App's secrets",
	Long: `The secrets command allows you to interact with your 
Keel App's secrets locally. This will allow you to add, remove, 
and list secrets that are stored in your cli config usually 
found at ~/.keel/config.yaml.`,
	Run: func(cmd *cobra.Command, args []string) {
		// list subcommands
		_ = cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(secretsCmd)
	secretsCmd.AddCommand(secretsListCmd)

	secretsListCmd.Flags().StringVar(&environment, "environment", "development", "The environment to use (default \"development\")")
}

var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all secrets for your Keel App",
	Long: `The list command will list all secrets that are
stored in your cli config usually found at ~/.keel/config.yaml.`,

	Run: func(cmd *cobra.Command, args []string) {
		program.Run(&program.Model{
			Mode:        program.ModeSecret,
			Environment: environment,
			Status:      program.StatusListSecrets,
			ProjectDir:  flagProjectDir,
		})
	},
}
