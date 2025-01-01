package main

import (
	"fmt"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/storage"
)

func handlerLogs() func(log routing.SensorLog) pubsub.AckType {
	return func(log routing.SensorLog) pubsub.AckType {

		err := storage.WriteLog(log)
		if err != nil {
			fmt.Printf("error writing log: %v\n", err)
			return pubsub.NackRequeue
		}

		return pubsub.Ack
	}
}
