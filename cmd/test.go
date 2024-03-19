package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/database"
	"github.com/teamkeel/keel/cmd/program"
	"github.com/teamkeel/keel/colors"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/schema/validation/errorhandling"
	"github.com/teamkeel/keel/testing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run tests",
	RunE: func(cmd *cobra.Command, args []string) error {
		// We still need a tracing provider for auditing and events,
		// even if the data is not being exported.
		otel.SetTracerProvider(sdktrace.NewTracerProvider())
		otel.SetTextMapPropagator(propagation.TraceContext{})

		// Only do bootstrap if no node_modules directory present
		_, err := os.Stat(filepath.Join(flagProjectDir, "node_modules"))
		if os.IsNotExist(err) {
			packageManager, err := resolvePackageManager(flagProjectDir, false)
			if err == promptui.ErrAbort {
				return nil
			}
			if err != nil {
				panic(err)
			}

			logPrefix := colors.Green("|").String()
			err = node.Bootstrap(
				flagProjectDir,
				node.WithPackageManager(packageManager),
				node.WithPackagesPath(flagNodePackagesPath),
				node.WithLogger(func(s string) {
					fmt.Println(logPrefix, s)
				}),
				node.WithOutputWriter(os.Stdout))
			if err != nil {
				return err
			}
		}

		// add "-test" suffix to the database so it doesn't clash with keel run db
		connInfo, err := database.Start(true, flagProjectDir+"-test")
		if err != nil {
			return err
		}

		secrets, err := program.LoadSecrets(flagProjectDir, "test")
		if err != nil {
			return err
		}

		err = testing.Run(context.Background(), &testing.RunnerOpts{
			Dir:            flagProjectDir,
			Pattern:        flagPattern,
			DbConnInfo:     connInfo,
			Secrets:        secrets,
			GenerateClient: false,
		})

		validationErrors := errorhandling.ValidationErrors{}
		exitError := &exec.ExitError{}

		switch {
		case errors.As(err, &validationErrors):
			fmt.Println("⚠️  Cannot run tests when schema contains errors. Run 'keel validate' to see error details.")
			fmt.Println("")
			return nil
		case err != nil && !errors.As(err, &exitError):
			return err
		default:
			return nil
		}
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	testCmd.Flags().StringVarP(&flagPattern, "pattern", "p", "(.*)", "pattern to isolate test")
	testCmd.Flags().StringVar(&flagPrivateKeyPath, "private-key-path", "", "path to the private key .pem file")

	if enabledDebugFlags == "true" {
		testCmd.Flags().StringVar(&flagNodePackagesPath, "node-packages-path", "", "path to local @teamkeel npm packages")
	}
}
