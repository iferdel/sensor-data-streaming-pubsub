package main

import (
	"fmt"

	"github.com/iferdel/treanteyes/internal/pubsub"
	"github.com/iferdel/treanteyes/internal/routing"
	"github.com/iferdel/treanteyes/internal/sensorlogic"
)

func handlerLogs() func(log routing.SensorLog) pubsub.AckType {
	return func(log routing.SensorLog) pubsub.AckType {

		err := sensorlogic.HandleLogs(log)
		if err != nil {
			fmt.Printf("error writing log: %v\n", err)
			return pubsub.NackRequeue
		}

		return pubsub.Ack
	}
}
