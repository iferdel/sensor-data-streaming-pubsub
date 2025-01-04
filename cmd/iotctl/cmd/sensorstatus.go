package cmd

import (
	"log"

	"github.com/iferdel/sensor-data-streaming-server/internal/storage"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get list of all sensors",
	Run: func(cmd *cobra.Command, args []string) {

		if conn != nil {
			defer conn.Close()
		}

		err := storage.GetSensor()
		if err != nil {
			log.Printf("could not retrieve sensors with get command: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
