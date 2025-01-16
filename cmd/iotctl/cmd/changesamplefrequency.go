package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/validation"
	"github.com/spf13/cobra"
)

var changeSampleFrequencyCmd = &cobra.Command{
	Use:   "changeSampleFrequency",
	Short: "Change the sample frequency [Hz] of a sensor",
	Run: func(cmd *cobra.Command, args []string) {

		if conn != nil {
			defer conn.Close()
		}

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

		newSampleFrequency, err := cmd.Flags().GetFloat64("changeSampleFrequency")
		if err != nil {
			log.Printf("error retrieving new sample frequency flag: %v", err)
			return
		}
		if newSampleFrequency <= 0 {
			fmt.Println("sample frequency must be a float greater than 0")
			return
		}

		err = pubsub.PublishGob(
			publishCh,                // amqp.Channel
			routing.ExchangeTopicIoT, // exchange
			fmt.Sprintf(routing.KeySensorCommandsFormat, sensorSerialNumber)+"."+"change_sample_frequency", // routing key
			routing.SensorCommandMessage{
				SerialNumber: sensorSerialNumber,
				Timestamp:    time.Now(),
				Command:      "changeSampleFrequency",
				Params: map[string]interface{}{
					"sampleFrequency": newSampleFrequency,
				},
			}, // value
		)
		if err != nil {
			log.Printf("could not publish change sample frequency command: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(changeSampleFrequencyCmd)
	changeSampleFrequencyCmd.Flags().StringP("sensor", "s", "", "sensorid")
	changeSampleFrequencyCmd.Flags().Float64P("changeSampleFrequency", "f", 1.0, "sample frequency")
}
