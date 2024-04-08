package bin

import (
	"ContainerMover/pkg/logger"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var (
	cfgFile string
	Info    bool
)

var rootCmd = &cobra.Command{
	Use:   "containerMover",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.containerMover/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&Info, "info", false, "logger ture for Info, false for Debug")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Find home directory.
	home := GetUserHomeDir()
	logFile := fmt.Sprintf("%s/.containerMover/containerMover.log", home)
	if !FileExist(home + "/.containerMover") {
		err := os.MkdirAll(home+"/.containerMover", os.ModePerm)
		if err != nil {
			fmt.Println("create default containerMover config dir failed, please create it by your self mkdir -p /root/.containerMover && touch /root/.containerMover/config.yaml")
		}
	}
	if Info {
		logger.Cfg(5, logFile)
	} else {
		logger.Cfg(6, logFile)
	}
}

func GetUserHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return home
}
