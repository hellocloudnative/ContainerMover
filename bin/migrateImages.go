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

	# migrate all images
	containerMover migrate images --src-type docker --dst-type containerd --namespace  --all
`
var migrateImagesCmd = &cobra.Command{
	Use:   "migrate images [flags] [image]",
	Short: "Migrate container images between different container runtimes",
	Long: `Migrate container images from a source runtime (e.g., Docker) to a destination runtime (e.g., Containerd).
The source and destination types can be "docker", "containerd", etc.`,
	Example: "containerMover migrate images --src-type docker --dst-type containerd myimage:latest",
	Run: func(cmd *cobra.Command, args []string) {
		if master.SrcType == "" || master.DstType == "" || master.Namespace == "" || (args[len(args)-1] == "images" && master.AllImages == false) {
			cmd.Help()
			os.Exit(1)
		}
		if master.AllImages == true {
			fmt.Println("Migrating all images from " + master.SrcType + " to " + master.DstType + " in namespace " + master.Namespace)
			if err := master.MigrateAllImages(master.SrcType, master.DstType, master.Namespace); err != nil {
				logger.Error("Migration failed: %v", err)
				os.Exit(1)
			}
			fmt.Println("Migration completed ")
			return
		}
		// 执行迁移操作
		imageName := args[len(args)-1]

		if err := master.MigrateImage(master.SrcType, master.DstType, imageName, master.Namespace); err != nil {
			logger.Error("Migration failed: %v", err)
			os.Exit(1)
		}

		fmt.Println("Migration completed ")
	},
}

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(migrateImagesCmd)
	}
	migrateImagesCmd.Flags().StringVar(&master.SrcType, "src-type", "docker", "The source container runtime type")
	migrateImagesCmd.Flags().StringVar(&master.DstType, "dst-type", "containerd", "The destination container runtime type")
	migrateImagesCmd.Flags().StringVar(&master.Namespace, "namespace", "k8s.io", "The namespace where the container images are located")
	migrateImagesCmd.Flags().BoolVarP(&master.AllImages, "all", "A", false, "Migrate all images in the namespace")
}
