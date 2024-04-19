package master

import (
	"ContainerMover/pkg/logger"
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"sync"
)

// MigrateAllImages concurrently migrates all images from a source runtime to a destination runtime.
func MigrateAllImages(srcType, dstType, namespace string) error {
	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %v", err)
	}
	defer dockerCli.Close()

	containerdCli, err := containerd.New("/run/containerd/containerd.sock", containerd.WithDefaultNamespace(namespace))
	if err != nil {
		return fmt.Errorf("failed to create Containerd client: %v", err)
	}
	defer containerdCli.Close()

	images, err := dockerCli.ImageList(context.Background(), types.ImageListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list Docker images: %v", err)
	}

	done := make(chan bool)
	errs := make(chan error, len(images))

	var wg sync.WaitGroup

	for _, img := range images {
		wg.Add(1)
		go func(img types.ImageSummary) {
			defer wg.Done()
			for _, tag := range img.RepoTags {
				err := MigrateImage(srcType, dstType, tag, namespace)
				if err != nil {
					errs <- fmt.Errorf("failed to migrate image %s: %v", tag, err)
					return
				}
				logger.Info("Successfully migrated image %s", tag)
			}
			done <- true
		}(img)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	for _ = range images {
		select {
		case err := <-errs:
			return err
		case <-done:
		}
	}

	logger.Info("All images migrated successfully")
	return nil
}
