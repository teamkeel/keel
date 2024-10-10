package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/program"
	"github.com/teamkeel/keel/colors"
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
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		if !lo.Contains([]string{"development", "test"}, flagEnvironment) {
			fmt.Printf("  %s Invalid --env %s: must be one of 'development' or 'test'\n\n", colors.Red("✘"), colors.Orange(flagEnvironment))
			fmt.Printf("  %s If you are trying to view secrets for a deployed environment then for Keel hosted projects\n", colors.Orange("|"))
			fmt.Printf("  %s you need to do this in the console and for self-hosted projects you need to use the\n", colors.Orange("|"))
			fmt.Printf("  %s `keel deploy secrets list` command.\n\n", colors.Orange("|"))
			os.Exit(1)
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
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		if !lo.Contains([]string{"development", "test"}, flagEnvironment) {
			fmt.Printf("  %s Invalid --env %s: must be one of 'development' or 'test'\n\n", colors.Red("✘"), colors.Orange(flagEnvironment))
			fmt.Printf("  %s If you are trying to set a secret for a deployed environment then for Keel hosted projects\n", colors.Orange("|"))
			fmt.Printf("  %s you need to do this in the console and for self-hosted projects you need to use the\n", colors.Orange("|"))
			fmt.Printf("  %s `keel deploy secrets set` command.\n\n", colors.Orange("|"))
			os.Exit(1)
		}

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
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		if !lo.Contains([]string{"development", "test"}, flagEnvironment) {
			fmt.Printf("  %s Invalid --env %s: must be one of 'development' or 'test'\n\n", colors.Red("✘"), colors.Orange(flagEnvironment))
			fmt.Printf("  %s If you are trying to remove a secret for a deployed environment then for Keel hosted projects\n", colors.Orange("|"))
			fmt.Printf("  %s you need to do this in the console and for self-hosted projects you need to use the\n", colors.Orange("|"))
			fmt.Printf("  %s `keel deploy secrets delete` command.\n\n", colors.Orange("|"))
			os.Exit(1)
		}

		err := program.RemoveSecret(flagProjectDir, flagEnvironment, key)
		if err != nil {
			return program.RenderError(err)
		}

		program.RenderSuccess(fmt.Sprintf("Secrets updated for environment %s", flagEnvironment))

		return nil
	},
}
