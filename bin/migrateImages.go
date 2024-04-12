package bin

import (
	"ContainerMover/master"
	"ContainerMover/pkg/logger"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var exampleInit = `
	# docker migrate to containerd
	containerMover  migrate  images --src-type docker --dst-type containerd  --namespace A
	
	# docker  migrate to isulad
	containerMover migrate images --src-type docker --dst-type isulad  --namespace B

	# containerd migrate to docker
    containerMover migrate images --src-type containerd --dst-type docker --namespace C
`
var migrateImagesCmd = &cobra.Command{
	Use:   "migrate images [flags] [image]",
	Short: "Migrate container images between different container runtimes",
	Long: `Migrate container images from a source runtime (e.g., Docker) to a destination runtime (e.g., Containerd).
The source and destination types can be "docker", "containerd", etc.`,
	Example: "containerMover migrate images --src-type docker --dst-type containerd myimage:latest",
	Run: func(cmd *cobra.Command, args []string) {
		if master.SrcType == "" || master.DstType == "" || master.Namespace == "" || len(args) == 0 {
			cmd.Help()
			os.Exit(1)
		}
		// 执行迁移操作
		imageName := args[len(args)-1]
		if err := master.MigrateImage(master.SrcType, master.DstType, imageName, master.Namespace); err != nil {
			logger.Error("Migration failed: %v", err)
			os.Exit(1)
		}
		// 迁移成功
		fmt.Printf("Image %s migrated from %s to %s successfully.\n", imageName, master.SrcType, master.DstType)
	},
}

func init() {
	rootCmd.AddCommand(migrateImagesCmd)
	migrateImagesCmd.Flags().StringVar(&master.SrcType, "src-type", "docker", "The source container runtime type")
	migrateImagesCmd.Flags().StringVar(&master.DstType, "dst-type", "containerd", "The destination container runtime type")
	migrateImagesCmd.Flags().StringVar(&master.Namespace, "namespace", "k8s.io", "The namespace where the container images are located")
}
