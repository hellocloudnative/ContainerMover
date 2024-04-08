package bin

import (
	"ContainerMover/master"
	"ContainerMover/pkg/logger"
	"github.com/spf13/cobra"
	"os"
)

var exampleInit = `
	# docker migrate to containerd
	containerMover  migrate  images --src-type docker --dst-type containerd   A
	
	# docker  migrate to isulad
	containerMover migrate images --src-type docker --dst-type isulad B

	# containerd migrate to docker
    containerMover migrate images --src-type containerd --dst-type docker C
`
var initCmd = &cobra.Command{
	Use:     "migrate",
	Short:   "migrate your container",
	Long:    `containerMover  migrate  images --src-type docker --dst-type containerd  A`,
	Example: exampleInit,
	Run: func(cmd *cobra.Command, args []string) {
		c := &master.ContainerMoverConfig{}
		// 没有重大错误可以直接保存配置. 但是apiservercertsans为空. 但是不影响用户 clean
		// 如果用户指定了配置文件,并不使用--master, 这里就不dump, 需要使用load获取配置文件了.
		if cfgFile != "" && len(master.Nodes) == 0 {
			err := c.Load(cfgFile)
			if err != nil {
				logger.Error("load cfgFile %s err: %q", cfgFile, err)
				os.Exit(1)
			}
		} else {
			c.Dump(cfgFile)
		}
		master.Migrate()
		c.Dump(cfgFile)
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		// 使用了cfgFile 就不进行preRun了
		if cfgFile == "" && master.ExitInitCase() {
			cmd.Help()
			os.Exit(master.ErrorExitOSCase)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
