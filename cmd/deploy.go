package cmd

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"context"
	"os"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/manifoldco/promptui"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/colors"
	"github.com/teamkeel/keel/deploy"
	"github.com/teamkeel/keel/runtime"
)

var flagDeployRuntimeBinary string
var flagDeployLogsSince time.Duration
var flagDeployLogsStart string

var deployCommand = &cobra.Command{
	Use:   "deploy",
	Short: "Self-host your Keel app on AWS",
	Run: func(cmd *cobra.Command, args []string) {
		// list subcommands
		_ = cmd.Help()
	},
}

type ValidateDeployFlagsResult struct {
	ProjectDir    string
	Env           string
	RuntimeBinary string
}

var envRegexp = regexp.MustCompile(`^[a-z\-0-9]+$`)

func validateDeployFlags() *ValidateDeployFlagsResult {
	// ensure environment is set and valid
	if flagEnvironment == "" {
		fmt.Println("You must specify an environment using the --env flag")
		return nil
	}
	if flagEnvironment == "development" || flagEnvironment == "test" {
		fmt.Println("--env cannot be 'development' or 'test' as these are reserved for 'keel run' and 'keel test' respectively")
		return nil
	}
	if !envRegexp.MatchString(flagEnvironment) {
		fmt.Println("--env can only contain lower-case letters, dashes, and numbers")
		return nil
	}

	// ensure project dir is absolute
	absProjectDir, err := filepath.Abs(flagProjectDir)
	if err != nil {
		panic(err)
	}

	v := runtime.GetVersion()
	runtimeBinary := fmt.Sprintf("https://github.com/teamkeel/keel/releases/download/v%s/runtime-lambda_%s_linux_amd64.tar.gz", v, v)

	if flagDeployRuntimeBinary != "" {
		runtimeBinary = flagDeployRuntimeBinary
		if !strings.HasPrefix(flagDeployRuntimeBinary, "http") {
			runtimeBinary, err = filepath.Abs(runtimeBinary)
			if err != nil {
				panic(err)
			}
		}
	}

	return &ValidateDeployFlagsResult{
		ProjectDir:    absProjectDir,
		Env:           flagEnvironment,
		RuntimeBinary: runtimeBinary,
	}
}

var deployBuildCommand = &cobra.Command{
	Use:   "build",
	Short: "Build the resources that are needed to deploy your app to your AWS account",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		defer panicHandler()

		validated := validateDeployFlags()
		if validated == nil {
			os.Exit(1)
		}

		_, err := deploy.Build(map[string]time.Time{}, context.Background(), &deploy.BuildArgs{
			ProjectRoot:   validated.ProjectDir,
			Env:           validated.Env,
			RuntimeBinary: validated.RuntimeBinary,
		})
		if err != nil {
			os.Exit(1)
		}
	},
}

var deployUpCommand = &cobra.Command{
	Use:   "up",
	Short: "Deploy your app into your AWS account for the given --env",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		defer panicHandler()

		validated := validateDeployFlags()
		if validated == nil {
			os.Exit(1)
		}

		err := deploy.Run(context.Background(), &deploy.RunArgs{
			Action:        deploy.UpAction,
			ProjectRoot:   validated.ProjectDir,
			Env:           validated.Env,
			RuntimeBinary: validated.RuntimeBinary,
		})
		if err != nil {
			os.Exit(1)
		}
	},
}

var deployRemoveCommand = &cobra.Command{
	Use:   "remove",
	Short: "Remove all resources in AWS for the given --env",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		defer panicHandler()

		validated := validateDeployFlags()
		if validated == nil {
			os.Exit(1)
		}

		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("This command will permanently delete all resources in your AWS account for the environment '%s', including your database.", validated.Env),
			IsConfirm: true,
		}

		_, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrAbort {
				fmt.Println("| Aborting...")
				return
			}
			panic(err)
		}

		err = deploy.Run(context.Background(), &deploy.RunArgs{
			Action:        deploy.RemoveAction,
			ProjectRoot:   validated.ProjectDir,
			Env:           validated.Env,
			RuntimeBinary: validated.RuntimeBinary,
		})
		if err != nil {
			os.Exit(1)
		}
	},
}

var deployLogsCommand = &cobra.Command{
	Use:   "logs",
	Short: "View logs from a deployed environment",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		defer panicHandler()

		validated := validateDeployFlags()
		if validated == nil {
			os.Exit(1)
		}

		if flagDeployLogsSince != 0 && flagDeployLogsStart != "" {
			fmt.Println("only one of --since and --start can be provided")
			os.Exit(1)
		}

		startTime := time.Now()
		if flagDeployLogsSince > 0 {
			startTime = startTime.Add(-flagDeployLogsSince)
		} else if flagDeployLogsStart != "" {
			var err error
			startTime, err = time.Parse(time.DateTime, flagDeployLogsStart)
			if err != nil {
				fmt.Println("--start has invalid value, must be in format 'YYYY-MM-DD HH:MM:SS'")
				os.Exit(1)
			}
		}

		//nolint:staticcheck
		err := deploy.StreamLogs(context.Background(), &deploy.StreamLogsArgs{
			ProjectRoot: validated.ProjectDir,
			Env:         validated.Env,
			StartTime:   startTime,
		})
		//nolint:staticcheck
		if err != nil {
			os.Exit(1)
		}
	},
}

var deploySecretsCommand = &cobra.Command{
	Use:   "secrets",
	Short: "Manage secrets for self-hosted Keel apps",
	Run: func(cmd *cobra.Command, args []string) {
		// list subcommands
		_ = cmd.Help()
	},
}

var deploySecretsSetCommand = &cobra.Command{
	Use:   "set MY_KEY 'my-value' --env my-env",
	Args:  cobra.ExactArgs(2),
	Short: "Set a secret in your AWS account for the given environment",
	Run: func(cmd *cobra.Command, args []string) {
		validated := validateDeployFlags()
		if validated == nil {
			os.Exit(1)
		}

		err := deploy.SetSecret(context.Background(), &deploy.SetSecretArgs{
			ProjectRoot: validated.ProjectDir,
			Env:         validated.Env,
			Key:         args[0],
			Value:       args[1],
		})
		if err != nil {
			os.Exit(1)
		}

		fmt.Printf("  %s Secret %s set\n", deploy.IconTick, colors.Orange(args[0]).String())
	},
}

var deploySecretsGetCommand = &cobra.Command{
	Use:   "get MY_KEY --env my-env",
	Args:  cobra.ExactArgs(1),
	Short: "Get a secret from your AWS account for the given environment",
	Run: func(cmd *cobra.Command, args []string) {
		validated := validateDeployFlags()
		if validated == nil {
			os.Exit(1)
		}

		p, err := deploy.GetSecret(context.Background(), &deploy.GetSecretArgs{
			ProjectRoot: validated.ProjectDir,
			Env:         validated.Env,
			Key:         args[0],
		})
		if err != nil {
			os.Exit(1)
		}

		fmt.Println(*p.Value)
	},
}

var deploySecretsListCommand = &cobra.Command{
	Use:   "list --env my-env",
	Args:  cobra.NoArgs,
	Short: "List the secrets set in your AWS account for the given environment",
	Run: func(cmd *cobra.Command, args []string) {
		validated := validateDeployFlags()
		if validated == nil {
			os.Exit(1)
		}

		params, err := deploy.ListSecrets(context.Background(), &deploy.ListSecretsArgs{
			ProjectRoot: validated.ProjectDir,
			Env:         validated.Env,
		})
		if err != nil {
			os.Exit(1)
		}

		var rows []table.Row
		keyLengths := []int{}
		valueLengths := []int{}
		for _, p := range params {
			parts := strings.Split(*p.Name, "/")
			name := parts[len(parts)-1]
			value := *p.Value

			// We don't show internal Keel secrets in this command, just user secrets
			if strings.HasPrefix(name, "KEEL_") {
				continue
			}

			keyLengths = append(keyLengths, len(name))
			valueLengths = append(valueLengths, len(value))

			rows = append(rows, table.Row{parts[len(parts)-1], *p.Value})
		}

		columns := []table.Column{
			{Title: "Name", Width: lo.Max(keyLengths)},
			{Title: "Value", Width: lo.Max(valueLengths)},
		}

		t := table.New(
			table.WithColumns(columns),
			table.WithRows(rows),
			table.WithHeight(len(rows)),
		)
		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			Bold(false)
		s.Selected = s.Selected.
			Foreground(lipgloss.NoColor{}).
			Bold(false)
		s.Cell = s.Cell.
			Foreground(colors.HighlightWhiteBright)

		t.SetStyles(s)

		secretsStyle := lipgloss.NewStyle().BorderStyle(lipgloss.NormalBorder())

		fmt.Println(secretsStyle.Render(t.View()))
	},
}

var deploySecretsDeleteCommand = &cobra.Command{
	Use:   "delete MY_KEY --env my-env",
	Args:  cobra.ExactArgs(1),
	Short: "Delete a secret set in your AWS account for the given environment",
	Run: func(cmd *cobra.Command, args []string) {
		validated := validateDeployFlags()
		if validated == nil {
			os.Exit(1)
		}

		err := deploy.DeleteSecret(context.Background(), &deploy.DeleteSecretArgs{
			ProjectRoot: validated.ProjectDir,
			Env:         validated.Env,
			Key:         args[0],
		})
		if err != nil {
			os.Exit(1)
		}

		fmt.Printf("  %s secret %s deleted\n", deploy.IconTick, colors.Orange(args[0]).String())
	},
}

func init() {
	rootCmd.AddCommand(deployCommand)

	deployCommand.AddCommand(deployBuildCommand)
	deployBuildCommand.Flags().StringVar(&flagEnvironment, "env", "", "The environment to build for e.g. staging or production")
	deployBuildCommand.Flags().StringVar(&flagDeployRuntimeBinary, "runtime-binary", "", "")

	deployCommand.AddCommand(deployUpCommand)
	deployUpCommand.Flags().StringVar(&flagEnvironment, "env", "", "The environment to deploy e.g. staging or production")
	deployUpCommand.Flags().StringVar(&flagDeployRuntimeBinary, "runtime-binary", "", "")

	deployCommand.AddCommand(deployRemoveCommand)
	deployRemoveCommand.Flags().StringVar(&flagEnvironment, "env", "", "The environment to remove e.g. staging or production")
	deployRemoveCommand.Flags().StringVar(&flagDeployRuntimeBinary, "runtime-binary", "", "")

	deployCommand.AddCommand(deployLogsCommand)
	deployLogsCommand.Flags().StringVar(&flagEnvironment, "env", "", "The environment to view logs for e.g. staging or production")
	deployLogsCommand.Flags().DurationVar(&flagDeployLogsSince, "since", 0, "--since 1h")
	deployLogsCommand.Flags().StringVar(&flagDeployLogsStart, "start", "", "--start '2024-11-01 09:00:00'")

	deployCommand.AddCommand(deploySecretsCommand)

	deploySecretsCommand.AddCommand(deploySecretsSetCommand)
	deploySecretsSetCommand.Flags().StringVar(&flagEnvironment, "env", "", "The environment to set the secret for e.g. staging or production")

	deploySecretsCommand.AddCommand(deploySecretsGetCommand)
	deploySecretsGetCommand.Flags().StringVar(&flagEnvironment, "env", "", "The environment to get the secret for e.g. staging or production")

	deploySecretsCommand.AddCommand(deploySecretsListCommand)
	deploySecretsListCommand.Flags().StringVar(&flagEnvironment, "env", "", "The environment to get the secret for e.g. staging or production")

	deploySecretsCommand.AddCommand(deploySecretsDeleteCommand)
	deploySecretsDeleteCommand.Flags().StringVar(&flagEnvironment, "env", "", "The environment to delete the secret from e.g. staging or production")
}
