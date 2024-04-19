package master

import (
	"ContainerMover/pkg/logger"
	"context"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/containerd/containerd"
	"github.com/docker/docker/client"
	"io"
	"time"
)

type progressReader struct {
	io.Reader
	bar *pb.ProgressBar
}

func (r *progressReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if r.bar.Current()+int64(n) <= r.bar.Total() {
		r.bar.Add(n)
	} else {
		r.bar.SetTotal(r.bar.Total())
	}
	return n, err
}

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

		bar := pb.New(int(img.VirtualSize)).Set(pb.Bytes, true).SetWidth(80)
		bar.Start()
		startTime := time.Now()
		switch dstType {
		case "containerd":
			progressReader := &progressReader{Reader: tarStream, bar: bar}

			ctx := context.Background()
			importOpts := []containerd.ImportOpt{containerd.WithIndexName(imageName)}

			_, err := containerdCli.Import(ctx, progressReader, importOpts...)
			if err != nil {
				return fmt.Errorf("failed to import image to Containerd: %v", err)
			}
			bar.Finish()
			elapsedTime := time.Since(startTime)
			logger.Info("Image %s migrated from Docker to Containerd successfully in %s.", imageName, elapsedTime)

		case "other-runtime":
			fmt.Printf("Migration to %s is not supported yet.\n", dstType)
		default:
			return fmt.Errorf("unsupported destination runtime: %s", dstType)
		}
	default:
		return fmt.Errorf("unsupported source runtime: %s", srcType)
	}

	return nil
}
