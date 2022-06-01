package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/formatter"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/validation"
)

type runCommand struct {
	outputFormatter *formatter.Output
}

// TODO - many opportunities to DRY this up alongside the validate command.

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run your Keel App locally",
	RunE:  commandImplementation,
}

func commandImplementation(cmd *cobra.Command, args []string) error {
	c := &runCommand{
		outputFormatter: formatter.New(os.Stdout),
	}
	switch outputFormat {
	case string(formatter.FormatJSON):
		c.outputFormatter.SetOutput(formatter.FormatJSON, os.Stdout)
	default:
		c.outputFormatter.SetOutput(formatter.FormatText, os.Stdout)
	}

	schema := schema.Schema{}
	var err error

	c.outputFormatter.Write("Reading your schema(s)")

	switch {
	case inputFile != "":
		_, err = schema.MakeFromFile(inputFile)
	default:
		_, err = schema.MakeFromDirectory(inputDir)
	}

	if err != nil {
		errs, ok := err.(validation.ValidationErrors)
		if ok {
			return c.outputFormatter.Write(errs.Errors)
		} else {
			return fmt.Errorf("error making schema: %v", err)
		}
	}

	c.outputFormatter.Write("Starting PostgreSQL")
	bringUpPostgres()

	return nil
}

func init() {
	rootCmd.AddCommand(runCmd)
	defaultDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("os.Getwd() errored: %v", err))
	}
	runCmd.Flags().StringVarP(&inputDir, "dir", "d", defaultDir, "schema directory to run")
	runCmd.Flags().StringVarP(&inputFile, "file", "f", "", "schema file to run")
	runCmd.Flags().StringVarP(&outputFormat, "output", "o", "console", "output format (console, json)")
}

func bringUpPostgres() error {
	ctx := context.Background()
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	imageName := "postgres"

	out, err := dockerClient.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	defer out.Close()
	io.Copy(os.Stdout, out)

	containerConfig := &container.Config{
		Image: imageName,
		Env:   []string{"POSTGRES_PASSWORD=admin123"},
	}
	resp, err := dockerClient.ContainerCreate(ctx, containerConfig, nil, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
	return nil
}
