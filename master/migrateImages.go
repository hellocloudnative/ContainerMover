package master

import (
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/docker/docker/client"
	"log"
)

// MigrateImage migrates an image from a source runtime to a destination runtime.
func MigrateImage(srcType string, dstType string, imageName, namspace string) error {
	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %v", err)
	}

	containerdCli, err := containerd.New("/run/containerd/containerd.sock", containerd.WithDefaultNamespace(namspace))
	if err != nil {
		return fmt.Errorf("failed to create Containerd client: %v", err)
	}
	defer containerdCli.Close()

	switch srcType {
	case "docker":
		img, _, err := dockerCli.ImageInspectWithRaw(context.Background(), imageName)
		if err != nil {
			return fmt.Errorf("failed to inspect Docker image: %v", err)
		}
		tarStream, err := dockerCli.ImageSave(context.Background(), []string{img.ID})
		if err != nil {
			return fmt.Errorf("failed to save Docker image: %v", err)
		}
		defer tarStream.Close()
		switch dstType {
		case "containerd":
			ctx := context.Background()
			importOpts := []containerd.ImportOpt{containerd.WithIndexName("docker.io/library/" + imageName)}
			_, err := containerdCli.Import(ctx, tarStream, importOpts...)

			if err != nil {
				return fmt.Errorf("failed to import image to Containerd: %v", err)
			}
			log.Printf("Image %s migrated from Docker to Containerd successfully.\n", imageName)
		case "other-runtime":
			log.Printf("Migration to %s is not supported yet.\n", dstType)
		default:
			return fmt.Errorf("unsupported destination runtime: %s", dstType)
		}

	default:
		return fmt.Errorf("unsupported source runtime: %s", srcType)
	}

	return nil
}
