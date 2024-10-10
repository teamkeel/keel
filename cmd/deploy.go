package cmd

import (
	"fmt"
	"time"

	"context"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/colors"
	"github.com/teamkeel/keel/deploy"
)

var flagDeployRuntimeBinaryURL string
var flagDeployLogsSince time.Duration
var flagDeployLogsStart string

// secretsCmd represents the secrets command
var deployCommand = &cobra.Command{
	Use:   "deploy",
	Short: "For Self-hosting your Keel app on AWS",
	Run: func(cmd *cobra.Command, args []string) {
		// list subcommands
		_ = cmd.Help()
	},
}

var deployUpCommand = &cobra.Command{
	Use:   "up",
	Short: "Deploy your app into your AWS account for the given --env",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		defer panicHandler()

		if flagEnvironment == "" {
			fmt.Println("You must specify an environment using the --env flag")
			os.Exit(1)
		}

		events, done := makeOutputChannel()
		defer func() {
			<-done
		}()

		// We don't need to handle the returned error as it will have been sent in the events channel
		_ = deploy.Run(context.Background(), &deploy.RunArgs{
			Action:      deploy.UpAction,
			ProjectRoot: flagProjectDir,
			Env:         flagEnvironment,
			Events:      events,

			// TODO: if not set this should be a URL to the artifact in the Github release
			RuntimeBinaryURL: flagDeployRuntimeBinaryURL,
		})
		return
	},
}

var deployDestroyCommand = &cobra.Command{
	Use:   "destroy",
	Short: "Remove all resources in AWS for the given --env",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		defer panicHandler()

		if flagEnvironment == "" {
			fmt.Println("You must specify an environment using the --env flag")
			os.Exit(1)
		}

		prompt := promptui.Prompt{
			Label:     "Running this command will permanently delete all resources in your AWS account, including your database.",
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

		events, done := makeOutputChannel()
		defer func() {
			<-done
		}()

		// We don't need to handle the returned error as it will have been sent in the events channel
		_ = deploy.Run(context.Background(), &deploy.RunArgs{
			Action:      deploy.DestroyAction,
			ProjectRoot: flagProjectDir,
			Env:         flagEnvironment,
			Events:      events,

			// TODO: if not set this should be a URL to the artifact in the Github release
			RuntimeBinaryURL: flagDeployRuntimeBinaryURL,
		})
		return
	},
}

var deployLogsCommand = &cobra.Command{
	Use:   "logs",
	Short: "View logs from a deployed environment",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		defer panicHandler()

		if flagEnvironment == "" {
			fmt.Println("You must specify an environment using the --env flag")
			os.Exit(1)
		}

		if flagDeployLogsSince != 0 && flagDeployLogsStart != "" {
			fmt.Println("Only one of --since and --start can be provided, not both")
			os.Exit(1)
		}

		// events, done := makeOutputChannel()
		// defer func() {
		// 	<-done
		// }()

		startTime := time.Now()
		if flagDeployLogsSince > 0 {
			startTime = startTime.Add(-flagDeployLogsSince)
		} else if flagDeployLogsStart != "" {
			var err error
			startTime, err = time.Parse(time.DateTime, flagDeployLogsStart)
			if err != nil {
				fmt.Println("Invalid --start flag, must be in format YYYY-MM-DD HH:MM::SS")
				os.Exit(1)
			}
		}

		err := deploy.StreamLogs(context.Background(), &deploy.StreamLogsArgs{
			ProjectRoot: flagProjectDir,
			Env:         flagEnvironment,
			StartTime:   startTime,
			// Events:      events,
		})
		if err != nil {
			panic(err)
		}
	},
}

func makeOutputChannel() (chan deploy.Output, chan bool) {
	evts := make(chan deploy.Output, 0)
	done := make(chan bool, 0)

	go func() {
		for event := range evts {
			if event.Heading != "" {
				fmt.Println("\n", colors.Heading(event.Heading).String())
				continue
			}

			switch event.Icon {
			case deploy.OutputIconTick:
				fmt.Print(colors.Green("✔").String())
			case deploy.OutputIconCross:
				fmt.Print(colors.Red("❌").String())
			case deploy.OutputIconPipe:
				fmt.Print("|")
			default:
				fmt.Print("|")
			}

			fmt.Printf(" %s\n", colors.White(event.Message).String())
			if event.Error != nil {
				fmt.Println("")
				fmt.Println("Error:")
				fmt.Println(" |", colors.Gray(event.Error.Error()).String())
				fmt.Println("")
			}
		}
		fmt.Println("")
		done <- true
	}()

	return evts, done
}

func init() {
	rootCmd.AddCommand(deployCommand)
	deployCommand.AddCommand(deployUpCommand)
	deployCommand.AddCommand(deployDestroyCommand)
	deployCommand.AddCommand(deployLogsCommand)
	deployCommand.PersistentFlags().StringVar(&flagEnvironment, "env", "", "")
	deployCommand.PersistentFlags().StringVar(&flagDeployRuntimeBinaryURL, "runtime-binary-url", "", "")

	deployLogsCommand.Flags().DurationVar(&flagDeployLogsSince, "since", 0, "--since 1h")
	deployLogsCommand.Flags().StringVar(&flagDeployLogsStart, "start", "", "--start '2024-11-01 09:00:00'")
}
