package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"github.com/teamkeel/keel/cmd/formatter"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/schema/validation"

	"github.com/fsnotify/fsnotify"
)

type runCommand struct {
	outputFormatter *formatter.Output
}

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
	return c.doTheWork()
}

func (c *runCommand) doTheWork() error {
	var err error
	if _, err = c.makeProtoFromSchemaFiles(); err != nil {
		return err
	}

	c.outputFormatter.Write("Starting PostgreSQL")
	bringUpPostgres()

	directoryWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating schema change watcher: %v", err)
	}
	defer directoryWatcher.Close()

	// All the action is triggered by reacting to changes being made to the input schema
	// files, so we set that up and then block this thread.
	handler := NewSchemaChangedHandler()
	go c.reactToSchemaChanges(directoryWatcher, handler)

	err = directoryWatcher.Add(inputDir)
	if err != nil {
		return fmt.Errorf("error specifying directory to schema watcher: %v", err)
	}

	c.outputFormatter.Write(fmt.Sprintf("Waiting for a schema file to change in %s ...\n", inputDir))

	ch := make(chan bool)
	<-ch

	// Todo - work out what resource clean up is required here.

	return nil
}

func init() {
	rootCmd.AddCommand(runCmd)
	defaultDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("os.Getwd() errored: %v", err))
	}
	runCmd.Flags().StringVarP(&inputDir, "dir", "d", defaultDir, "schema directory to run")
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
	// What to do about this output? In its naive form it is not part of the CLI Run command contract,
	// but does a good job of showing the progress of this slow-running step.

	// io.Copy(os.Stdout, out)

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

func (c *runCommand) reactToSchemaChanges(watcher *fsnotify.Watcher, handler *SchemaChangedHandler) {
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				fmt.Printf("XXXX this signals that the watcher event channel got closed. No known stimulii yet.\n")
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				nameOfFileThatChanged := event.Name
				if err := handler.Handle(nameOfFileThatChanged); err != nil {
					panic(fmt.Errorf("error handling schema change event: %v", err))
				}
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				fmt.Printf("XXXX this signifies the watcher error channel got closed. No known stimulli yet.\n")
				return
			}
			fmt.Printf("XXXX error received on watcher error channel: %v\n", err)
		}
	}
}

type SchemaChangedHandler struct {
}

func NewSchemaChangedHandler() *SchemaChangedHandler {
	return &SchemaChangedHandler{}
}

func (h *SchemaChangedHandler) Handle(schemaThatHasChanged string) error {
	fmt.Printf("XXXX handler fired because this file: %s, changed\n", schemaThatHasChanged)
	return nil
}

func (c *runCommand) makeProtoFromSchemaFiles() (proto *proto.Schema, err error) {
	c.outputFormatter.Write("Reading your schema(s)")
	schema := schema.Schema{}
	// todo - inputDir is a cmd package-global variable (because it is a CLI command flag), but we
	// should introduce a pass-by-value copy to pass down the call stack.
	proto, err = schema.MakeFromDirectory(inputDir)
	if err != nil {
		errs, ok := err.(validation.ValidationErrors)
		if ok {
			return nil, c.outputFormatter.Write(errs.Errors)
		} else {
			return nil, fmt.Errorf("error making schema: %v", err)
		}
	}
	return proto, nil
}
