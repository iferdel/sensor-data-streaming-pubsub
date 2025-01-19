package cmd

import (
	"fmt"
	"log"

	"github.com/iferdel/sensor-data-streaming-server/internal/validation"
	"github.com/spf13/cobra"
)

var awakeCmd = &cobra.Command{
	Use:   "awake",
	Short: "Awake sensor from sleep and restart generating/sending data",
	Run: func(cmd *cobra.Command, args []string) {

		sensorSerialNumber, err := cmd.Flags().GetString("sensor")
		if err != nil {
			log.Printf("error retrieving sensorid flag: %v", err)
			return
		}
		if sensorSerialNumber == "" {
			log.Printf("sensor serial number cannot be empty")
			return
		}
		if !validation.HasValidCharacters(sensorSerialNumber) {
			log.Printf("sensor serial number not valid")
			return
		}

		all, err := cmd.Flags().GetBool("all")
		if err != nil {
			log.Printf("error retrieving all flag: %v", err)
			return
		}

		if all {
			fmt.Println("sending awake command to all sensors -- not yet implemented")
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(awakeCmd)
	awakeCmd.Flags().StringP("sensor", "s", "", "sensorid")
	awakeCmd.Flags().BoolP("all", "a", false, "awake all sensors")
}
