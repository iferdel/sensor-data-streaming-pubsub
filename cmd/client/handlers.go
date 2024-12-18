package main

import (
	"fmt"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/sensorlogic"
)

func handlerCommand(sensorState *sensorlogic.SensorState) func(cm routing.CommandMessage) pubsub.AckType {
	return func(cm routing.CommandMessage) pubsub.AckType {
		switch cm.Command {
		case "sleep": // convert to constants command strings
			handleSleep(sensorState)
		}
		return pubsub.Ack
	}
}

func handleSleep(sensorState *sensorlogic.SensorState) {
	if sensorState.IsSleep {
		fmt.Println("sensor is already in a sleep state")
		return
	}
	sensorState.IsSleep = true
	fmt.Println("sensor is set to sleep")
}
