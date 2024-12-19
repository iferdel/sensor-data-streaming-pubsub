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
		defer conn.Close()

		sensorSerialNumber, err := cmd.Flags().GetString("sensorid")
		if err != nil {
			log.Printf("error retrieving sensorid flag: %v", err)
			return
		}
		fmt.Println("sending sleep command to sensor", sensorSerialNumber)
		err = pubsub.PublishGob(
			publishCh,                // amqp.Channel
			routing.ExchangeTopicIoT, // exchange
			fmt.Sprintf(routing.BindKeySensorCommandFormat, sensorSerialNumber), // routing key
			routing.CommandMessage{
				SensorName: sensorSerialNumber,
				Timestamp:  time.Now(),
				Command:    "sleep",
				Params:     nil,
			}, // value
		)
		if err != nil {
			log.Printf("could not publish sleep command: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(sleepCmd)
	sleepCmd.Flags().StringP("sensorid", "s", "", "sensorid")
}
