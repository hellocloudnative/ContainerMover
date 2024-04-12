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
	Short: "Move containers between different runtimes",
	Long: `ContainerMover is a CLI tool that facilitates the migration of containers
from one runtime to another. It supports various source and destination runtimes
such as Docker, Containerd, and others.`,
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
	rootCmd.PersistentFlags().BoolVar(&Info, "info", false, "set logger level to Info (default is Debug)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Find home directory.
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error finding home directory:", err)
		os.Exit(1)
	}

	logFile := fmt.Sprintf("%s/.containerMover/containerMover.log", home)
	if !FileExist(home + "/.containerMover") {
		err := os.MkdirAll(home+"/.containerMover", os.ModePerm)
		if err != nil {
			fmt.Printf("Failed to create config directory: %v\n", err)
			fmt.Println("Please create it manually with the command: mkdir -p /root/.containerMover && touch /root/.containerMover/config.yaml")
			os.Exit(1)
		}
	}

	// Set the logger configuration based on the Info flag.
	if Info {
		logger.Cfg(5, logFile) // Assuming 5 is the log level for Info.
	} else {
		logger.Cfg(6, logFile) // Assuming 6 is the log level for Debug.
	}
}

func FileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
