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
stored in your cli config usually found at ~/.keel/config.yaml.`,

	RunE: func(cmd *cobra.Command, args []string) error {
		secrets, err := program.LoadSecrets(flagProjectDir)
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
	Long:  "The set command will set a secret for your Keel App.",

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return program.RenderError(errors.New("Not enough arguments, please provide a key and value"))
		}

		key := args[0]
		value := args[1]

		err := program.SetSecret(flagProjectDir, key, value)
		if err != nil {
			return program.RenderError(err)
		}

		program.RenderSuccess(fmt.Sprintf("Secret %s set", key))

		return nil
	},
}

var secretsRemoveCmd = &cobra.Command{
	Use:   "remove <env> <key>",
	Short: "Remove a secret for your Keel App",
	Long:  "The remove command will remove a secret for your Keel App",

	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return program.RenderError(errors.New("Not enough arguments, please provide a key"))
		}

		key := args[0]

		err := program.RemoveSecret(flagProjectDir, key)
		if err != nil {
			return program.RenderError(err)
		}

		program.RenderSuccess("Secrets updated")

		return nil
	},
}
