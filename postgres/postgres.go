// package postgres exists to keep all-things postgres in one place.
package postgres

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/samber/lo"
)

// BringUpPostgresLocally spins up a PostgreSQL server locally and returns
// a connection to it.
//
// It deploys it with Docker.
// It pulls the postres docker image if it is not already available locally.
// It leaves the default superuser name untouched "postgres".
// It sets the password for that user to "postgres".
func BringUpPostgresLocally() error {
	if err := bringUpContainer(); err != nil {
		return err
	}
	// todo establish and return the connection
	return nil
}

func bringUpContainer() error {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	fmt.Printf("Checking if postgres image already present... ")
	postgresImage, err := findPostresImageLocally(dockerClient)
	if err != nil {
		return err
	}

	switch {
	case postgresImage != nil:
		fmt.Printf("it is\n")
	default:
		fmt.Printf("it is not, so fetching it\n")
		// todo set a timeout for the pull
		// todo - consider if doing the image pull syncronously is ok.
		// todo - consider showing some output from the pull operation (first return value)
		reader, err := dockerClient.ImagePull(context.Background(), postgresImageName, types.ImagePullOptions{})
		if err != nil {
			return fmt.Errorf("error pulling postgres image: %v", err)
		}
		defer reader.Close()
		// ImagePull is async - and the suggested protocol for waiting for it to complete is
		// to read from the returned reader, until you reach EOF.
		awaitReadCompletion(reader)
		fmt.Printf("Fetched ok\n")
	}

	fmt.Printf("Checking if postgres container exists... ")
	postgresContainer, err := findPostgresContainer(dockerClient)
	if err != nil {
		return err
	}

	switch {
	case postgresContainer != nil:
		fmt.Printf("it does\n")
	default:
		fmt.Printf("it does not\n")
		fmt.Printf("Creating container... ")

		containerConfig := &container.Config{
			Image: postgresImageName,
			Env: []string{
				"POSTGRES_PASSWORD=postgres",
			},
		}
		if _, err := dockerClient.ContainerCreate(
			context.Background(),
			containerConfig,
			nil,
			nil,
			nil,
			keelPostgresContainerName); err != nil {
			return fmt.Errorf("error creating postgres container: %v", err)
		}
		postgresContainer, _ = findPostgresContainer(dockerClient)
		fmt.Printf("created\n")
	}

	// See if container is running
	fmt.Printf("Checking if postgres container is already running... ")
	isRunning, err := isContainerRunning(dockerClient, postgresContainer)
	if err != nil {
		return err
	}

	switch {
	case isRunning:
		fmt.Printf("it is\n")
	default:
		fmt.Printf("it is not\n")
		err := dockerClient.ContainerStart(
			context.Background(),
			postgresContainer.ID,
			types.ContainerStartOptions{})
		if err != nil {
			return fmt.Errorf("error starting postgres container: %v", err)
		}
		fmt.Printf("Started it\n")
	}
	return nil
}

func findPostresImageLocally(dockerClient *client.Client) (*types.ImageSummary, error) {
	images, err := dockerClient.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return nil, err
	}
	searchFor := strings.Join([]string{postgresImageName, postgresTag}, ":")
	for _, image := range images {
		tags := image.RepoTags
		if lo.Contains(tags, searchFor) {
			return &image, nil
		}
	}
	return nil, nil
}

func findPostgresContainer(dockerClient *client.Client) (container *types.Container, err error) {
	listOptions := types.ContainerListOptions{
		All: true,
	}
	containers, err := dockerClient.ContainerList(context.Background(), listOptions)
	if err != nil {
		return nil, err
	}
	for _, c := range containers {
		if lo.Contains(c.Names, "/"+keelPostgresContainerName) {
			return &c, nil
		}
	}
	return nil, nil
}

func isContainerRunning(dockerClient *client.Client, container *types.Container) (bool, error) {
	containerJSON, err := dockerClient.ContainerInspect(context.Background(), container.ID)
	if err != nil {
		return false, err
	}
	return containerJSON.State.Running, nil
}

func awaitReadCompletion(r io.Reader) {
	// Consuming the output in (max) 1000 byte chunks gives us circa
	// 80 read cycles - and we output a dot for each of them to show progress.
	buf := make([]byte, 1000)
	for {
		_, err := r.Read(buf)
		fmt.Printf(".")
		if err == io.EOF {
			break
		}
	}
}

const postgresImageName string = "postgres"
const postgresTag string = "latest"

const keelPostgresContainerName string = "keel-postgres"
