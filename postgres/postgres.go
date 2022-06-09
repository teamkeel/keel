// package postgres exists to keep all-things postgres in one place.
package postgres

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"database/sql"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq"
	"github.com/samber/lo"
)

// BringUpPostgresLocally spins up a PostgreSQL server locally and returns
// a connection to it.
//
// It is the client's responsibility to call db.Close() on the returned
// connection when done with it.
//
// It deploys it with Docker.
// It pulls the postres docker image if it is not already available locally.
// It leaves the default superuser name untouched "postgres".
// It sets the password for that user to "postgres".
func BringUpPostgresLocally() (*sql.DB, error) {
	if err := bringUpContainer(); err != nil {
		return nil, err
	}
	connection, err := establishConnection()
	if err != nil {
		return nil, err
	}
	return connection, nil
}

// StopThePostgresContainer stops the postgres container - having checked first
// that such a container exists, and it is running.
func StopThePostgresContainer() error {
	fmt.Printf("Stopping the postgres container... ")
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	container, err := findPostgresContainer(dockerClient)
	if err != nil {
		return err
	}
	if container == nil {
		return nil
	}
	running, err := isContainerRunning(dockerClient, container)
	if err != nil {
		return err
	}
	if !running {
		return nil
	}
	stopTimeout := time.Duration(5 * time.Second)
	// Note that ContainerStop() gracefully stops the container by choice, but then forcibly after the timeout.
	err = dockerClient.ContainerStop(context.Background(), container.ID, &stopTimeout)
	if err != nil {
		return err
	}
	fmt.Printf("Stopped\n")
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
			makeHostConfig(),
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

// awaitReadCompletion reads from the given reader until it reaches EOF.
// It's used in the context of waiting for a docker image to be fetched, and
// is the method used in the docker SDK to wait for the fetch to be complete.
// We exploit it also to emit a growing row of dot characters to indicate
// progress.
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

func makeHostConfig() *container.HostConfig {
	portBinding := nat.PortBinding{
		HostIP:   "",
		HostPort: "5432", // todo could be problematic to have this hard coded
	}
	portMap := nat.PortMap{
		nat.Port("5432/tcp"): []nat.PortBinding{portBinding},
	}
	hostConfig := &container.HostConfig{
		PortBindings: portMap,
	}
	return hostConfig
}

func establishConnection() (*sql.DB, error) {
	psqlInfo := "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

const postgresImageName string = "postgres"
const postgresTag string = "latest"

const keelPostgresContainerName string = "keel-postgres"
