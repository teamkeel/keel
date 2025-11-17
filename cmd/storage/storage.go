package storage

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/docker/docker/api/types/container"
	dockerContainer "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	dockerVolume "github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/samber/lo"
	"github.com/teamkeel/keel/util"
)

type ConnectionInfo struct {
	Host      string
	Port      string
	UIPort    string
	Bucket    string
	AccessKey string
	SecretKey string
	Region    string
}

const minioImageName string = "minio/minio"
const minioImageTag string = "latest"
const minioAccessKey string = "keelstorage"
const minioSecretKey string = "keelstorage"
const minioRegion string = "us-east-1"

const keelStorageContainerName string = "keel-run-storage"
const keelStorageVolumeName string = "keel-storage-volume"

// Start spins up a dockerised Minio service locally
//
// It is the client's responsibility to call db.Close() on the returned
// connection when done with it.
//
func Start(projectDirectory string) (*ConnectionInfo, error) {
	// We need a dockerClient (proxy) to "drive" Docker using the SDK.
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	// Create the volume if it doesn't exist
	if err := createVolumeIfNotExists(dockerClient); err != nil {
		return nil, err
	}

	// Create the container for THIS run.
	containerID, port, uiPort, err := createContainer(dockerClient)
	if err != nil {
		return nil, err
	}

	// Start the container running. (This is where it chooses the port to serve on)
	if err := startContainer(dockerClient, containerID); err != nil {
		return nil, err
	}
	conn := ConnectionInfo{
		Host:      "127.0.0.1",
		Port:      *port,
		UIPort:    *uiPort,
		AccessKey: minioAccessKey,
		SecretKey: minioSecretKey,
		Region:    minioRegion,
	}

	// Wait for MinIO to be ready before creating bucket
	if err := waitForMinioReady(context.Background(), &conn); err != nil {
		return nil, err
	}

	bucketName, err := createProjectBucket(context.Background(), projectDirectory, &conn)
	if err != nil {
		return nil, err
	}
	conn.Bucket = bucketName

	return &conn, nil
}

// Stop stops the minio container - having checked first that such a container exists, and it is running.
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

// fetchImageIfNecessary goes off to fetch the official Postgres Docker image,
// if it is not already stored in Docker's local Image Repository.
func fetchImageIfNecessary(dockerClient *client.Client) error {
	postgresImage, err := findImage(dockerClient)
	if err != nil {
		return err
	}

	if postgresImage == nil {
		fmt.Println("Pulling minio Docker image...")
		reader, err := dockerClient.ImagePull(context.Background(), minioImageName+":"+minioImageTag, image.PullOptions{})
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

// removeContainer removes the Minio container if there already is one present.
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

// createContainer creates a fresh Minio container.
//
// It first removes the previously created container if one exists.
// It configures it to serve on a port that is free RIGHT NOW. (favouring 6000)
// And it mounts our long-lived Docker storage volume into the container's
// file system - at the location that Minio uses to store the files.
func createContainer(dockerClient *client.Client) (
	containerID string,
	port *string,
	uiPort *string,
	err error,
) {
	if err := fetchImageIfNecessary(dockerClient); err != nil {
		return "", nil, nil, err
	}
	if err := removeContainer(dockerClient); err != nil {
		return "", nil, nil, err
	}

	containerConfig := &container.Config{
		Image: minioImageName + ":" + minioImageTag,
		Env: []string{
			"MINIO_ROOT_USER=" + minioAccessKey,
			"MINIO_ROOT_PASSWORD=" + minioSecretKey,
		},
		Cmd: []string{"server", "/data", "--console-address", ":9001"},
		ExposedPorts: nat.PortSet{
			"9000/tcp": struct{}{},
			"9001/tcp": struct{}{},
		},
	}

	hostPort, err := util.GetFreePort("8010")
	if err != nil {
		return "", nil, nil, err
	}
	hostUIPort, err := util.GetFreePort("8011")
	if err != nil {
		return "", nil, nil, err
	}

	createdInfo, err := dockerClient.ContainerCreate(
		context.Background(),
		containerConfig,
		&container.HostConfig{
			PortBindings: nat.PortMap{
				nat.Port("9000/tcp"): []nat.PortBinding{{HostPort: hostPort}},
				nat.Port("9001/tcp"): []nat.PortBinding{{HostPort: hostUIPort}},
			},
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeVolume,
					Source: keelStorageVolumeName,
					Target: "/data",
				},
			},
		},
		nil, // network config
		nil, // platform config
		keelStorageContainerName)

	if err != nil {
		return "", nil, nil, err
	}

	return createdInfo.ID, &hostPort, &hostUIPort, nil
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

// findImage looks up the minio Docker Image we require in the local
// Docker Image Resistry and returns its metadata. It it is not registered,
// it signals this by returning nil metadata.
func findImage(dockerClient *client.Client) (*image.Summary, error) {
	images, err := dockerClient.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return nil, err
	}
	searchFor := strings.Join([]string{minioImageName, minioImageTag}, ":")
	for _, image := range images {
		tags := image.RepoTags
		if lo.Contains(tags, searchFor) {
			return &image, nil
		}
	}
	return nil, nil
}

// findContainer obtains a reference to the Minio container we make, if one exists.
// If it cannot find it, it then returns container as nil.
func findContainer(dockerClient *client.Client) (container *dockerContainer.Summary, err error) {
	containers, err := dockerClient.ContainerList(context.Background(), dockerContainer.ListOptions{
		All: true,
	})
	if err != nil {
		return nil, err
	}
	for _, c := range containers {
		if lo.Contains(c.Names, "/"+keelStorageContainerName) {
			return &c, nil
		}
	}
	return nil, nil
}

// isContainerRunning returns true if the given container is currently running.
func isContainerRunning(dockerClient *client.Client, container *dockerContainer.Summary) (bool, error) {
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
	buf := make([]byte, 4000)
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

// generateBucketName generates a unique but deterministic bucket name using a
// hash of the project's working directory
// For example: keel_48f77af86bffe7cdbb44308a70d11f8b.
func generateBucketName(projectDirectory string) (string, error) {
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

	return fmt.Sprintf("%x", md5.Sum([]byte(projectDirectory))), nil
}

// waitForMinioReady waits for MinIO to be ready to accept connections by retrying
// ListBuckets calls until successful or timeout is reached.
func waitForMinioReady(ctx context.Context, conn *ConnectionInfo) error {
	endpoint := fmt.Sprintf("http://%s:%s", conn.Host, conn.Port)
	s3Client := s3.NewFromConfig(aws.Config{
		BaseEndpoint: &endpoint,
		Credentials:  credentials.NewStaticCredentialsProvider(minioAccessKey, minioSecretKey, ""),
		Region:       minioRegion,
	})

	maxRetries := 30
	retryDelay := 200 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		_, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
		if err == nil {
			// MinIO is ready
			return nil
		}

		// Wait before retrying
		time.Sleep(retryDelay)
	}

	return fmt.Errorf("timeout waiting for MinIO to be ready")
}

// createProjectBucket will create a new bucket for the given project. If the bucket already exists, it's name will be returned.
func createProjectBucket(ctx context.Context, projectDirectory string, conn *ConnectionInfo) (string, error) {
	bucketName, err := generateBucketName(projectDirectory)
	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf("http://%s:%s", conn.Host, conn.Port)
	s3Client := s3.NewFromConfig(aws.Config{
		BaseEndpoint: &endpoint,
		Credentials:  credentials.NewStaticCredentialsProvider(minioAccessKey, minioSecretKey, ""),
		Region:       minioRegion,
	})

	list, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{
		Prefix: &bucketName,
	})
	if err != nil {
		return "", err
	}

	for _, b := range list.Buckets {
		if b.Name != nil && *b.Name == bucketName {
			// bucket already exists
			return bucketName, nil
		}
	}

	_, err = s3Client.CreateBucket(context.Background(), &s3.CreateBucketInput{
		Bucket: &bucketName,
	})
	if err != nil {
		return "", err
	}

	return bucketName, nil
}

// createVolumeIfNotExists creates the Docker volume we'll persist the storage data on,
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
		dockerVolume.CreateOptions{Name: keelStorageVolumeName})
	if err != nil {
		return err
	}

	return nil
}

// findVolume returns the volume we use to persist the storage data on, if it
// exists. If it does not yet exist it returns the volume as nil.
func findVolume(dockerClient *client.Client) (volume *dockerVolume.Volume, err error) {
	volList, err := dockerClient.VolumeList(context.Background(), dockerVolume.ListOptions{Filters: filters.Args{}})
	if err != nil {
		return nil, err
	}
	for _, vol := range volList.Volumes {
		if vol.Name == keelStorageVolumeName {
			return vol, nil
		}
	}
	return nil, nil
}
