package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/spf13/cobra"
)

var sleepCmd = &cobra.Command{
	Use:   "sleep",
	Short: "Stop sensor from generating/sending more data",
	Run: func(cmd *cobra.Command, args []string) {

		if conn != nil {
			defer conn.Close()
		}

		sensorSerialNumber, err := cmd.Flags().GetString("sensor")
		if err != nil {
			log.Printf("error retrieving sensorid flag: %v", err)
			return
		}

		all, err := cmd.Flags().GetBool("all")
		if err != nil {
			log.Printf("error retrieving all flag: %v", err)
			return
		}

		if all {
			fmt.Println("sending sleep command to all sensors -- not yet implemented")
			return
		}

		fmt.Println("sending sleep command to sensor", sensorSerialNumber)
		err = pubsub.PublishGob(
			publishCh,                // amqp.Channel
			routing.ExchangeTopicIoT, // exchange
			fmt.Sprintf(routing.KeySensorCommandsFormat, sensorSerialNumber)+"."+"sleep", // routing key
			routing.SensorCommandMessage{
				SerialNumber: sensorSerialNumber,
				Timestamp:    time.Now(),
				Command:      "sleep",
				Params:       nil,
			}, // value
		)
		if err != nil {
			log.Printf("could not publish sleep command: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(sleepCmd)
	sleepCmd.Flags().StringP("sensor", "s", "", "sensorid")
	sleepCmd.Flags().BoolP("all", "a", false, "sleep all sensors")
}
