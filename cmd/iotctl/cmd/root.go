package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "iotctl",
	Short: "CLI Tool for Managing IoT Sensors via REST API",
	Long: `This CLI tool allows you to manage resources (sensors) 
It allows the use of keywords to alter the behaviour of the available sensors in the cluster. 
Every command has some flags such as sensor id or parameters related to the command itself.`,
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
