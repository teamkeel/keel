// package postgres exists to keep all-things postgres in one place.
package postgres

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
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
		return fmt.Errorf("error creating docker client: %v", err)
	}

	ctx := context.Background()
	// todo make the pull conditional on it not being available locally.
	// todo set a timeout for the pull
	// todo - consider if doing the image pull syncronously is ok.
	// todo - consider showing some output from the pull operation (first return value)
	out, err := dockerClient.ImagePull(ctx, "postgres", types.ImagePullOptions{})
	if err != nil {
		return fmt.Errorf("error pulling postgres image: %v", err)
	}
	defer out.Close()

	containerConfig := &container.Config{
		Image: "postgres",
		Env: []string{
			"POSTGRES_PASSWORD=postgres",
		},
	}
	resp, err := dockerClient.ContainerCreate(ctx, containerConfig, nil, nil, nil, "")
	if err != nil {
		return fmt.Errorf("error creating postgres container: %v", err)
	}

	if err := dockerClient.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("error starting postgres container: %v", err)
	}
	return nil
}
