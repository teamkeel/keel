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
	secretsCmd.PersistentFlags().StringVarP(&flagEnvironment, "env", "e", "development", "environment")
}

var secretsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all secrets for your Keel App",
	Long: `The list command will list all secrets that are
stored in your cli config usually found at ~/.keel/config.yaml.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return program.RenderError(errors.New("Too many arguments"))
		}

		secrets, err := program.LoadSecrets(flagProjectDir, flagEnvironment)
		if err != nil {
			return program.RenderError(err)
		}
		if len(secrets) == 0 {
			return program.RenderError(errors.New("No secrets found"))
		}

		program.RenderSuccess(fmt.Sprintf("Listing secrets for environment: %s", flagEnvironment))
		fmt.Println(program.RenderSecrets(secrets))

		return nil
	},
}

var secretsSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a secret for your Keel App",
	Long:  "The set command will set a secret for your Keel App. The default environment is development.",

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return program.RenderError(errors.New("Not enough arguments, please provide a key and value"))
		}
		if len(args) > 2 {
			return program.RenderError(errors.New("Too many arguments"))
		}

		key := args[0]
		value := args[1]

		err := program.SetSecret(flagProjectDir, flagEnvironment, key, value)
		if err != nil {
			return program.RenderError(err)
		}

		program.RenderSuccess(fmt.Sprintf("Secret %s set for environment %s", key, flagEnvironment))

		return nil
	},
}

var secretsRemoveCmd = &cobra.Command{
	Use:   "remove <key>",
	Short: "Remove a secret for your Keel App",
	Long:  "The remove command will remove a secret for your Keel App. The default environment is development.",

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return program.RenderError(errors.New("Not enough arguments, please provide a key"))
		}
		if len(args) > 1 {
			return program.RenderError(errors.New("Too many arguments"))
		}

		key := args[0]

		err := program.RemoveSecret(flagProjectDir, flagEnvironment, key)
		if err != nil {
			return program.RenderError(err)
		}

		program.RenderSuccess(fmt.Sprintf("Secrets updated for environment %s", flagEnvironment))

		return nil
	},
}
