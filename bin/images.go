package bin

import (
	"ContainerMover/master"
	"ContainerMover/pkg/logger"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

var exampleInit = `
	# docker migrate to containerd
	containerMover  images --src-type docker --dst-type containerd  --namespace A
	
	# docker  migrate to isulad
	containerMover images --src-type docker --dst-type isulad  --namespace B

	# containerd migrate to docker
    containerMover images --src-type containerd --dst-type docker --namespace C

	# migrate all images
	containerMover images --src-type docker --dst-type containerd --namespace  --all

	# Migrate images listed in a file from Docker to Containerd
	containerMover images --src-type docker --dst-type containerd --image-list imagelist.txt

	# docker migrate to  remote containerd
	containerMover images --src-type docker --dst-type containerd  --hosts 192.168.1.2, 192.168.1.3 --username root  --password 123456  A B
`
var migrateImagesCmd = &cobra.Command{
	Use:   "images",
	Short: "Migrate container images between different container runtimes",
	Long: `Migrate container images from a source runtime (e.g., Docker) to a destination runtime (e.g., Containerd).
The source and destination types can be "docker", "containerd", etc.`,
	Example: "containerMover migrate images --src-type docker --dst-type containerd myimage:latest",
	Run: func(cmd *cobra.Command, args []string) {
		if master.SrcType == "" || master.DstType == "" || master.Namespace == "" {
			cmd.Help()
			fmt.Println("Error: You must specify --src-type, --dst-type, and --namespace.")
			os.Exit(1)
		}

		var imageNames []string
		if master.ImageListFile != "" {
			content, err := ioutil.ReadFile(master.ImageListFile)
			if err != nil {
				logger.Error("Error reading image list file: %v", err)
				os.Exit(1)
			}
			imageNames = strings.Split(strings.TrimSpace(string(content)), "\n")
		} else if master.AllImages {
			if err := master.MigrateAllImages(master.SrcType, master.DstType, master.Namespace); err != nil {
				logger.Error("Migration failed: %v", err)
				os.Exit(1)
			}
			fmt.Println("All images have been migrated successfully.")
			return
		} else {
			imageNames = args
		}

		if len(imageNames) == 0 {
			fmt.Println("Error: You must provide image names, an image list file with --image-list, or use --all to migrate all images.")
			cmd.Help()
			os.Exit(1)
		}
		var wg sync.WaitGroup
		errs := make(chan error, len(master.Hosts)*len(imageNames))
		// 检查是否有远程主机地址提供
		if len(master.Hosts) > 0 {
			for _, host := range master.Hosts {
				wg.Add(len(imageNames))
				for _, imageName := range imageNames {
					go func(host, imageName string) {
						defer wg.Done()
						if err := master.MigrateImageRemotely(master.SrcType, master.DstType, master.Namespace, imageName, host, master.Username, master.Password); err != nil {
							errs <- fmt.Errorf("failed to migrate image %s to host %s: %v", imageName, host, err)
						} else {
							fmt.Printf("Successfully migrated image %s to host %s\n", imageName, host)
						}
					}(host, imageName)
				}
			}
		} else {
			for _, imageName := range imageNames {
				wg.Add(1)
				go func(imageName string) {
					defer wg.Done()
					if err := master.MigrateImage(master.SrcType, master.DstType, imageName, master.Namespace); err != nil {
						errs <- fmt.Errorf("failed to migrate image %s: %v", imageName, err)
					} else {
						fmt.Printf("Image %s has been migrated successfully from %s to %s.\n", imageName, master.SrcType, master.DstType)
					}
				}(imageName)
			}

		}
		go func() {
			wg.Wait()
			close(errs)
		}()

		for err := range errs {
			logger.Error("Migration error: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	if rootCmd != nil {
		rootCmd.AddCommand(migrateImagesCmd)
	}

	migrateImagesCmd.Flags().StringVar(&master.SrcType, "src-type", "docker", "Source runtime type (e.g., docker)")
	migrateImagesCmd.Flags().StringVar(&master.DstType, "dst-type", "containerd", "Destination runtime type (e.g., containerd)")
	migrateImagesCmd.Flags().StringVar(&master.Namespace, "namespace", "k8s.io", "Namespace where the container images are located")
	migrateImagesCmd.Flags().BoolVar(&master.AllImages, "all", false, "Migrate all images in the namespace")
	migrateImagesCmd.Flags().StringVar(&master.ImageListFile, "image-list", "", "File containing a list of image names to migrate, one per line")
	migrateImagesCmd.Flags().StringSliceVar(&master.Hosts, "hosts", nil, "Remote host addresses to migrate images to")
	migrateImagesCmd.Flags().StringVar(&master.Username, "username", "", "Username for authenticating with the hosts")
	migrateImagesCmd.Flags().StringVar(&master.Password, "password", "", "Password for authenticating with the hosts")
}
