package main

import (
	"fmt"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/sensorlogic"
)

func handlerCommand(cfg *Config, sensorState *sensorlogic.SensorState) func(cm routing.SensorCommandMessage) pubsub.AckType {
	return func(cm routing.SensorCommandMessage) pubsub.AckType {
		publishCh, err := cfg.rabbitConn.Channel()
		if err != nil {
			fmt.Println("error creating channel")
			return pubsub.NackRequeue
		}
		defer publishCh.Close()

		switch cm.Command {
		case "sleep": // convert to constants command strings
			sensorState.HandleSleep()
		case "awake":
			sensorState.HandleAwake()
		case "changeSampleFrequency":
			newSampleFreq, err := sensorState.HandleChangeSampleFrequency(cm.Params)
			if err != nil {
				fmt.Printf("error while changing sample frequency: %v", err)
			}
			err = publishSensorLog(
				publishCh,
				routing.SensorLog{
					SerialNumber: sensorState.Sensor.SerialNumber,
					Timestamp:    time.Now(),
					Level:        "INFO",
					Message:      fmt.Sprintf("Sample frequency changed to %v [Hz]", newSampleFreq),
				},
			)
			if err != nil {
				fmt.Printf("Error publishing log: %s\n", err)
			}
		default:
			fmt.Println("not a valid command")
			return pubsub.NackDiscard
		}
		return pubsub.Ack
	}
}
