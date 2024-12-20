package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/spf13/cobra"
)

var awakeCmd = &cobra.Command{
	Use:   "awake",
	Short: "Awake sensor from sleep and restart generating/sending data",
	Run: func(cmd *cobra.Command, args []string) {

		if conn != nil {
			defer conn.Close()
		}

		sensorSerialNumber, err := cmd.Flags().GetString("sensor")
		if err != nil {
			log.Printf("error retrieving sensorid flag: %v", err)
			return
		}

		fmt.Println("sending awake command to sensor", sensorSerialNumber)
		err = pubsub.PublishGob(
			publishCh,                // amqp.Channel
			routing.ExchangeTopicIoT, // exchange
			fmt.Sprintf(routing.BindKeySensorCommandFormat, sensorSerialNumber), // routing key
			routing.CommandMessage{
				SensorName: sensorSerialNumber,
				Timestamp:  time.Now(),
				Command:    "awake",
				Params:     nil,
			}, // value
		)
		if err != nil {
			log.Printf("could not publish awake command: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(awakeCmd)
	awakeCmd.Flags().StringP("sensor", "s", "", "sensorid")
}
