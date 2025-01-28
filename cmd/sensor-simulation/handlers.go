package main

import (
	"github.com/iferdel/treanteyes/internal/pubsub"
	"github.com/iferdel/treanteyes/internal/routing"
	"github.com/iferdel/treanteyes/internal/sensorlogic"
)

func handlerCommand(sensorState *sensorlogic.SensorState) func(cm routing.SensorCommandMessage) pubsub.AckType {
	return func(cm routing.SensorCommandMessage) pubsub.AckType {

		switch cm.Command {
		case "sleep": // convert to constants command strings
			sensorState.HandleSleep()
		case "awake":
			sensorState.HandleAwake()
		case "changeSampleFrequency":
			sensorState.HandleChangeSampleFrequency(cm.Params)
		default:
			sensorState.LogsWarning <- "not a valid command"
			return pubsub.NackDiscard
		}
		return pubsub.Ack
	}
}
