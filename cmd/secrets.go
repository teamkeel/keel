package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/program"
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
}

var secretsListCmd = &cobra.Command{
	Use:   "list <env>",
	Short: "List all secrets for your Keel App",
	Long: `The list command will list all secrets that are
stored in your cli config usually found at ~/.keel/config.yaml. 
The default environment is development.`,

	Run: func(cmd *cobra.Command, args []string) {
		environment := "development"
		if len(args) > 0 {
			environment = args[0]
		}

		secrets, err := program.LoadSecrets(flagProjectDir, environment)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if len(secrets) == 0 {
			fmt.Println("No secrets found")
			os.Exit(0)
		}

		fmt.Println(program.RenderSecrets(secrets))
	},
}
