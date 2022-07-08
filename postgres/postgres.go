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

// You can specify if it should use the existing container, so as to retain its mounted
// volume, and thus stored data. This is good for the Run command. Conversely, the /runtime/gql/api test
// suite, needs to have a virgin container for each test case, to avoid conflicts with prior state.
//
// But it can also cope with a virgin run when even the required Docker image
// is not in the local docker registery.
//
// It sets the password for that user to "postgres".
// It sets the default database name to "keel"
func BringUpPostgresLocally(useFreshContainer bool) (*sql.DB, error) {
	if err := bringUpContainer(useFreshContainer); err != nil {
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
	return nil
}

func bringUpContainer(useFreshContainer bool) error {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	postgresImage, err := findPostresImageLocally(dockerClient)
	if err != nil {
		return err
	}

	if postgresImage == nil {
		reader, err := dockerClient.ImagePull(context.Background(), postgresImageName, types.ImagePullOptions{})
		if err != nil {
			return err
		}
		defer reader.Close()
		// ImagePull() is async - and the suggested protocol for waiting for it to complete is
		// to read from the returned reader, until you reach EOF.
		awaitReadCompletion(reader)
	}

	postgresContainer, err := findPostgresContainer(dockerClient)
	if err != nil {
		return err
	}

	// If we want to create a fresh container, but there is already one registered
	// with that name, we have to remove it first to be able to re create it.
	if postgresContainer != nil && useFreshContainer {
		if err := removeContainer(dockerClient, postgresContainer); err != nil {
			return err
		}
	}

	if postgresContainer == nil || useFreshContainer {
		containerConfig := &container.Config{
			Image: postgresImageName,
			Env: []string{
				"POSTGRES_PASSWORD=postgres",
				"POSTGRES_DB=keel",
			},
		}

		if _, err := dockerClient.ContainerCreate(
			context.Background(),
			containerConfig,
			makeHostConfig(),
			nil,
			nil,
			keelPostgresContainerName); err != nil {
			return err
		}
		postgresContainer, _ = findPostgresContainer(dockerClient)
	}

	// See if container is running
	isRunning, err := isContainerRunning(dockerClient, postgresContainer)
	if err != nil {
		return err
	}

	if !isRunning {
		err := dockerClient.ContainerStart(
			context.Background(),
			postgresContainer.ID,
			types.ContainerStartOptions{})
		if err != nil {
			return err
		}
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

func removeContainer(dockerClient *client.Client, container *types.Container) error {
	if err := dockerClient.ContainerRemove(
		context.Background(),
		container.ID,
		types.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         true,
		}); err != nil {
		return err
	}
	return nil
}

// awaitReadCompletion reads from the given reader until it reaches EOF.
// It's used in the context of waiting for a docker image to be fetched, and
// is the method used in the docker SDK to wait for the fetch to be complete.
// We exploit it also to emit a growing row of dot characters to indicate
// progress.
func awaitReadCompletion(r io.Reader) {
	// Consuming the output in N-byte chunks gives us circa
	// a friendly number of read cycles - good for outputting a progress dot "." for each of them.
	buf := make([]byte, 2000)
	for {
		_, err := r.Read(buf)
		fmt.Printf(".")
		if err != nil {
			if err != io.EOF {
				panic(fmt.Sprintf("error from read operation: %v", err))

			}
			return
		}
	}
}

func makeHostConfig() *container.HostConfig {
	portBinding := nat.PortBinding{
		HostIP:   "",
		HostPort: "5432",
	}
	portMap := nat.PortMap{
		nat.Port("5432/tcp"): []nat.PortBinding{portBinding},
	}
	hostConfig := &container.HostConfig{
		PortBindings: portMap,
	}
	return hostConfig
}

// establishConnection connects to the database, veryifies the connection and returns the connection.
// It makes a series of attempts over a small time span to give postgres the
// change to be ready.
func establishConnection() (*sql.DB, error) {
	psqlInfo := "host=localhost port=5432 user=postgres password=postgres dbname=keel sslmode=disable"
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	// Attempt to ping() the database at 250ms intervals a few times.
	var pingError error
	for i := 0; i < 10; i++ {
		if pingError = db.Ping(); pingError == nil {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	fmt.Printf("\n")
	if pingError != nil {
		return nil, fmt.Errorf("could not ping the database, despite several retries: %v", pingError)
	}
	return db, nil
}

const postgresImageName string = "postgres"
const postgresTag string = "latest"

const keelPostgresContainerName string = "keel-postgres"
