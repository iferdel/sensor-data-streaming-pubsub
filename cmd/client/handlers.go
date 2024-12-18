package main

import (
	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/sensorlogic"
)

func handlerCommand(sensorState *sensorlogic.SensorState) func(cm routing.CommandMessage) pubsub.AckType {
	return func(cm routing.CommandMessage) pubsub.AckType {
		switch cm.Command {
		case "sleep": // convert to constants command strings
			sensorState.HandleSleep()
		}
		return pubsub.Ack
	}
}
