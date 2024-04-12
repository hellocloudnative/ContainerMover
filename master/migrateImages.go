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
	// 创建 Docker 客户端
	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %v", err)
	}

	// 创建 Containerd 客户端
	containerdCli, err := containerd.New("/run/containerd/containerd.sock", containerd.WithDefaultNamespace(namspace))
	if err != nil {
		return fmt.Errorf("failed to create Containerd client: %v", err)
	}
	defer containerdCli.Close()

	// 根据源类型处理镜像迁移
	switch srcType {
	case "docker":
		// 从 Docker 保存镜像
		img, _, err := dockerCli.ImageInspectWithRaw(context.Background(), imageName)
		if err != nil {
			return fmt.Errorf("failed to inspect Docker image: %v", err)
		}
		tarStream, err := dockerCli.ImageSave(context.Background(), []string{img.ID})
		if err != nil {
			return fmt.Errorf("failed to save Docker image: %v", err)
		}
		defer tarStream.Close()
		// 根据目标类型处理镜像
		switch dstType {
		case "containerd":
			// 创建导入镜像的上下文
			ctx := context.Background()
			// 设置导入选项，包括镜像的索引名称
			importOpts := []containerd.ImportOpt{containerd.WithIndexName("docker.io/library/" + imageName)}
			_, err := containerdCli.Import(ctx, tarStream, importOpts...)

			if err != nil {
				return fmt.Errorf("failed to import image to Containerd: %v", err)
			}
			log.Printf("Image %s migrated from Docker to Containerd successfully.\n", imageName)
		case "other-runtime":
			// 这里添加其他目标运行时的镜像迁移逻辑
			log.Printf("Migration to %s is not supported yet.\n", dstType)
			// 可以根据需要返回错误或继续执行其他逻辑
		default:
			return fmt.Errorf("unsupported destination runtime: %s", dstType)
		}

	default:
		return fmt.Errorf("unsupported source runtime: %s", srcType)
	}

	return nil
}
