package cmd

import (
	"log"

	"github.com/iferdel/sensor-data-streaming-server/internal/storage"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete sensor from database (and all its data)",
	Run: func(cmd *cobra.Command, args []string) {

		if conn != nil {
			defer conn.Close()
		}

		sensorSerialNumber, err := cmd.Flags().GetString("sensor")
		if err != nil {
			log.Printf("error retrieving sensorid flag: %v", err)
			return
		}

		err = storage.DeleteSensor(sensorSerialNumber)
		if err != nil {
			log.Printf("could not delete sensor (%s) with get command: %v", sensorSerialNumber, err)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringP("sensor", "s", "", "sensorid")
}
