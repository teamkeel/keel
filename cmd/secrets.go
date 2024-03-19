package cmd

import (
	"errors"
	"fmt"

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
	secretsCmd.AddCommand(secretsSetCmd)
	secretsCmd.AddCommand(secretsRemoveCmd)
}

var secretsListCmd = &cobra.Command{
	Use:   "list <env>",
	Short: "List all secrets for your Keel App",
	Long: `The list command will list all secrets that are
stored in your cli config usually found at ~/.keel/config.yaml. 
The default environment is development.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		environment := "development"
		if len(args) > 0 {
			environment = args[0]
		}

		secrets, err := program.LoadSecrets(flagProjectDir, environment)
		if err != nil {
			return program.RenderError(err)
		}
		if len(secrets) == 0 {
			return program.RenderError(errors.New("No secrets found"))
		}

		fmt.Println(program.RenderSecrets(secrets))

		return nil
	},
}

var secretsSetCmd = &cobra.Command{
	Use:   "set <env> <key> <value>",
	Short: "Set a secret for your Keel App",
	Long:  "The set command will set a secret for your Keel App. The default environment is development.",

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 3 {
			return program.RenderError(errors.New("Not enough arguments, please provide an environment, key, and value"))
		}

		environment := args[0]
		key := args[1]
		value := args[2]

		err := program.SetSecret(flagProjectDir, environment, key, value)
		if err != nil {
			return program.RenderError(err)
		}

		program.RenderSuccess(fmt.Sprintf("Secret %s set for environment %s", key, environment))

		return nil
	},
}

var secretsRemoveCmd = &cobra.Command{
	Use:   "remove <env> <key>",
	Short: "Remove a secret for your Keel App",
	Long:  "The remove command will remove a secret for your Keel App. The default environment is development.",

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return program.RenderError(errors.New("Not enough arguments, please provide an environment and key"))
		}

		environment := args[0]
		key := args[1]

		err := program.RemoveSecret(flagProjectDir, environment, key)
		if err != nil {
			return program.RenderError(err)
		}

		program.RenderSuccess(fmt.Sprintf("Secrets updated for environment %s", environment))

		return nil
	},
}
