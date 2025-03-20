package database

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"database/sql"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	dockerVolume "github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/util"
)

// Start spins up a dockerised PostgreSQL server locally and returns
// a connection to it.
//
// It is the client's responsibility to call db.Close() on the returned
// connection when done with it.
//
// It creates a separate dedicated database for each of your projects - based
// on the project's directory name.

// You have to specify if you want the database contents for THIS project to be
// reset (cleared) or retained.
//
// You don't need to have a Postgres Docker image already available - because it will
// go and fetch one if necessary the first time.
//
// It creates and starts a fresh docker container each time in order to
// be able to serve Postgres on a port that is free on the host, RIGHT NOW.
// It favours port 5432 when possible.
//
// However it manages a "live forever" Docker Volume and tells Postgres
// to persist the database contents to that volume. Consequently the Volume
// and the data are long lived, while the container is ephemeral.
//
// It names the default database to be 'postgres' and sets the pg password
// also to "postgres".
//
// The connection info it returns refers to the project-specific database.
// The bool returned is true if the database was created, false if it already existed.
func Start(reset bool, projectDirectory string) (*db.ConnectionInfo, error, bool) {
	// We need a dockerClient (proxy) to "drive" Docker using the SDK.
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err, false
	}

	// We tell every Postgres container we launch and run, to use a long-lived,
	// external Docker Volume to store the database contents and metadata on. So
	// we take this opportunity to create that Volume, if it doesn't already
	// exist on this host's Docker environment. (i.e. only happens once).
	if err := createVolumeIfNotExists(dockerClient); err != nil {
		return nil, err, false
	}

	// Create the container for THIS run.
	// Note the returned connection is for the DEFAULT Postgres
	// database - NOT a project-specific database.
	containerID, serverConnectionInfo, err := createContainer(dockerClient)
	if err != nil {
		return nil, err, false
	}

	// Start the container running. (This is where it chooses the port to serve on)
	if err := startContainer(dockerClient, containerID); err != nil {
		return nil, err, false
	}

	// Calculate the deterministic and unique, project-specific database name
	projectDbName, err := generateDatabaseName(projectDirectory)
	if err != nil {
		return nil, err, false
	}

	projectDatabaseExists, err := doesDbExist(serverConnectionInfo, projectDbName)
	if err != nil {
		return nil, err, false
	}

	createdDb := false

	// Obey the mandate to clear the project-specific database if requested,
	// by DROP-ing that database.

	if projectDatabaseExists && reset {
		if err := dropDatabase(serverConnectionInfo, projectDbName); err != nil {
			return nil, err, false
		}
		projectDatabaseExists = false // We just removed it.
	}

	// Make sure the project database exists. It may never have existed yet, or we might
	// have just dropped it to do a reset.
	if !projectDatabaseExists {
		createdDb = true
		if err := createProjectDatabase(serverConnectionInfo, projectDbName); err != nil {
			return nil, err, false
		}
	}

	// We return a project-specific connectionInfo that points to the
	// project-specific database.
	projectConnectionInfo := serverConnectionInfo.WithDatabase(projectDbName)
	return projectConnectionInfo, nil, createdDb
}

// Stop stops the postgres container - having checked first
// that such a container exists, and it is running.
//
// This is no longer strictly necessary now, but it seems likely the user would prefer not
// to have needless containers running. In fact we could with this new architecture delete the
// container at this point. It get's deleted anyhow when you next run Keel run.
// But it leaving it as it was because its harmless.
func Stop() error {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	container, err := findContainer(dockerClient)
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
	stopTimeout := int((5 * time.Second).Seconds())
	// Note that ContainerStop() gracefully stops the container by choice, but then forcibly after the timeout.
	if err := dockerClient.ContainerStop(context.Background(), container.ID, dockerContainer.StopOptions{Timeout: &stopTimeout}); err != nil {
		return err
	}
	return nil
}

// fetchPostgresImageIfNecessary goes off to fetch the official Postgres Docker image,
// if it is not already stored in Docker's local Image Repository.
func fetchPostgresImageIfNecessary(dockerClient *client.Client) error {
	postgresImage, err := findImage(dockerClient)
	if err != nil {
		return err
	}

	if postgresImage == nil {
		fmt.Println("Pulling postgres Docker image...")
		reader, err := dockerClient.ImagePull(context.Background(), postgresImageName+":"+postgresTag, image.PullOptions{})
		if err != nil {
			return err
		}
		defer reader.Close()
		// ImagePull() is async - and the suggested protocol for waiting for it to complete is
		// to read from the returned reader, until you reach EOF.
		awaitReadCompletion(reader)
	}
	// Double check it worked.
	if _, err := findImage(dockerClient); err != nil {
		return err
	}
	return nil
}

// removeContainer removes the PG container if there already is one present.
func removeContainer(dockerClient *client.Client) error {
	containerMetadata, err := findContainer(dockerClient)
	if err != nil {
		return err
	}
	if containerMetadata == nil {
		return nil
	}
	if err := dockerClient.ContainerRemove(
		context.Background(),
		containerMetadata.ID,
		dockerContainer.RemoveOptions{
			RemoveVolumes: false,
			Force:         true,
		}); err != nil {
		return err
	}
	return nil
}

// createContainer creates a fresh PG container.
//
// It first removes the previously created container if one exists.
// It configures the default database to be named 'postgres'.
// It configures it to serve on a port that is free RIGHT NOW. (favouring 5432)
// And it mounts our long-lived Docker storage volume into the container's
// file system - at the location that Postgres uses to store the database contents.
func createContainer(dockerClient *client.Client) (
	containerID string,
	connInfo *db.ConnectionInfo, err error) {
	if err := fetchPostgresImageIfNecessary(dockerClient); err != nil {
		return "", nil, err
	}
	if err := removeContainer(dockerClient); err != nil {
		return "", nil, err
	}

	// Note we are delibarately leaving the Postgres DB storage file system location at its
	// default of '/var/lib/postgresql/data' (rather than setting the envvar PGDATA),
	// because that is where we have mounted the external storage volume to the container.
	containerConfig := &container.Config{
		Image: postgresImageName + ":" + postgresTag,
		Env: []string{
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_USER=postgres",
		},
	}

	port, err := util.GetFreePort("5432", "54321")
	if err != nil {
		return "", nil, err
	}

	hostConfig := newPortBindingAndVolumeMountConfig(port)

	createdInfo, err := dockerClient.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		nil, // network config
		nil, // platform config
		keelPostgresContainerName)

	if err != nil {
		return "", nil, err
	}
	connInfo = &db.ConnectionInfo{
		Username: "postgres",
		Password: "postgres",
		Host:     "127.0.0.1",
		Port:     port,
	}
	return createdInfo.ID, connInfo, nil
}

// startContainer starts the container with the given ID.
func startContainer(dockerClient *client.Client, containerID string) error {
	if err := dockerClient.ContainerStart(
		context.Background(),
		containerID,
		dockerContainer.StartOptions{}); err != nil {
		return err
	}
	return nil
}

// findImage looks up the Postgres Docker Image we require in the local
// Docker Image Resistry and returns its metadata. It it is not registered,
// it signals this by returning nil metadata.
func findImage(dockerClient *client.Client) (*image.Summary, error) {
	images, err := dockerClient.ImageList(context.Background(), image.ListOptions{})
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

// findContainer obtains a reference to the Postgres container we make, if one exists.
// If it cannot find it it returns container as nil.
func findContainer(dockerClient *client.Client) (container *types.Container, err error) {
	containers, err := dockerClient.ContainerList(context.Background(), dockerContainer.ListOptions{
		All: true,
	})
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

// isContainerRunning returns true if the given container is currently running.
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
			fmt.Printf("\n")
			return
		}
	}
}

// newPortBindingAndVolumeMountConfig makes a HostConfig object.
//
// This includes port mapping from the port the container will serve on in the
// Docker VPN to the given host port on the host - thus making the container
// accessible to the host's network.
// And it includes mounting our long-lived Docker storage volume at the
// file system location in the container that Docker uses to persist the database
// contents and metadata. The latter makes sure that the database contents are
// persisted beyond the life of any one container.
func newPortBindingAndVolumeMountConfig(hostPort string) *container.HostConfig {
	portBinding := nat.PortBinding{
		HostIP:   "",
		HostPort: hostPort,
	}
	portMap := nat.PortMap{
		nat.Port("5432/tcp"): []nat.PortBinding{portBinding},
	}
	pgStorageMount := mount.Mount{
		Type:   mount.TypeVolume,
		Source: keelPGVolumeName,
		Target: keelVolumeMountPath,
	}
	hostConfig := &container.HostConfig{
		PortBindings: portMap,
		Mounts:       []mount.Mount{pgStorageMount},
	}
	return hostConfig
}

// generateDatabaseName generates a unique but deterministic database name using a
// hash of the project's working directory
// For example: keel_48f77af86bffe7cdbb44308a70d11f8b
func generateDatabaseName(projectDirectory string) (string, error) {
	if strings.HasPrefix(projectDirectory, "~/") {
		home, _ := os.UserHomeDir()
		projectDirectory = filepath.Join(home, projectDirectory[2:])
	}

	// Ensure path is absolute and cleaned for determinism.
	projectDirectory, err := filepath.Abs(projectDirectory)
	if err != nil {
		return "", err
	}

	projectDirectory = strings.ToLower(projectDirectory)

	return fmt.Sprintf("keel_%x", md5.Sum([]byte(projectDirectory))), nil
}

// doesDbExist tells you if the database of the given name already exists.
func doesDbExist(serverConnectionInfo *db.ConnectionInfo, dbName string) (exists bool, err error) {
	server, err := connectAndWaitForDbServer(serverConnectionInfo)
	if err != nil {
		return false, err
	}

	result := server.QueryRow(
		fmt.Sprintf("SELECT COUNT(*) as count FROM pg_database WHERE datname = '%s'",
			dbName))
	var count int
	err = result.Scan(&count)
	// I've seen this indicate that the database does not exist in two different ways.
	// 1) Error: "no rows found"
	// 2) count == 0
	// Don't know why - so it checks for either case.
	if err != nil || count == 0 {
		return false, nil
	}
	return true, nil
}

// createProjectDatabase creates a database of the given name. It will return an error
// if the database already exists.
func createProjectDatabase(serverConnectionInfo *db.ConnectionInfo, dbToCreate string) error {
	server, err := connectAndWaitForDbServer(serverConnectionInfo)
	if err != nil {
		return err
	}
	_, err = server.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbToCreate))
	if err != nil {
		return err
	}
	return nil
}

// dropDatabase tells Postgres to drop the database of the given name.
func dropDatabase(serverConnectionInfo *db.ConnectionInfo, dbToDrop string) error {
	server, err := connectAndWaitForDbServer(serverConnectionInfo)
	if err != nil {
		return err
	}
	_, err = server.Exec(fmt.Sprintf("DROP DATABASE %s;", dbToDrop))
	if err != nil {
		return err
	}
	return nil
}

// connectAndWaitForDbServer connects to the database prescribed the given connection info,
// and waits for it to be ready before returning.
func connectAndWaitForDbServer(serverConnectionInfo *db.ConnectionInfo) (server *sql.DB, err error) {
	server, err = sql.Open("pgx/v5", serverConnectionInfo.String())
	if err != nil {
		return nil, err
	}

	// ping() the database until it is available.
	var pingError error
	for i := 0; i < 20; i++ {
		if pingError = server.Ping(); pingError == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	return server, pingError
}

// createVolume creates the Docker volume we'll persist the database(s) on,
// if it does not already exist.
func createVolumeIfNotExists(dockerClient *client.Client) error {
	vol, err := findVolume(dockerClient)
	if err != nil {
		return err
	}
	if vol != nil {
		return nil
	}
	_, err = dockerClient.VolumeCreate(
		context.Background(),
		dockerVolume.CreateOptions{Name: keelPGVolumeName})
	if err != nil {
		return nil
	}

	return nil
}

// findVolume returns the volume we use to persist the database on, if it
// exists. If it does not yet exist it returns the volume as nil.
func findVolume(dockerClient *client.Client) (volume *dockerVolume.Volume, err error) {
	volList, err := dockerClient.VolumeList(context.Background(), dockerVolume.ListOptions{Filters: filters.Args{}})
	if err != nil {
		return nil, err
	}
	for _, vol := range volList.Volumes {
		if vol.Name == keelPGVolumeName {
			return vol, nil
		}
	}
	return nil, nil
}

const postgresImageName string = "pgvector/pgvector"
const postgresTag string = "pg16"
const keelPostgresContainerName string = "keel-run-postgres"
const keelPGVolumeName string = "keel-pg-volume-v16"
const keelVolumeMountPath = `/var/lib/postgresql/data`
