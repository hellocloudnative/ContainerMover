package master

import (
	"ContainerMover/pkg/logger"
	"context"
	"fmt"
	"github.com/containerd/containerd"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// MigrateAllImages migrates all images from a source runtime to a destination runtime.
func MigrateAllImages(srcType, dstType, namespace string) error {
	// 创建 Docker 客户端
	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %v", err)
	}
	defer dockerCli.Close()

	// 创建 Containerd 客户端
	containerdCli, err := containerd.New("/run/containerd/containerd.sock", containerd.WithDefaultNamespace(namespace))
	if err != nil {
		return fmt.Errorf("failed to create Containerd client: %v", err)
	}
	defer containerdCli.Close()

	// 获取所有 Docker 镜像
	images, err := dockerCli.ImageList(context.Background(), types.ImageListOptions{All: true})
	if err != nil {
		return fmt.Errorf("failed to list Docker images: %v", err)
	}

	// 遍历所有镜像并迁移
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if err := MigrateImage(srcType, dstType, tag, namespace); err != nil {
				logger.Info("Failed to migrate image %s: %v", tag, err)
			} else {
				logger.Info("Successfully migrated image %s", tag)
			}
		}
	}

	// 如果所有镜像都成功迁移，则打印成功消息
	if count := len(images); count == 0 {
		logger.Info("No images to migrate.")
	} else {
		logger.Info("All %d images migrated successfully.", count)
	}

	return nil
}
