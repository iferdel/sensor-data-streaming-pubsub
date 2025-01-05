package main

import (
	"fmt"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/storage"
)

func handlerSensorRegistry() func(s routing.Sensor) pubsub.AckType {
	return func(s routing.Sensor) pubsub.AckType {
		// placeholder
		fmt.Println("==========================================")

		err := storage.WriteSensor(s.SerialNumber)
		if err != nil {
			fmt.Printf("error writing sensor instance: %v\n", err)
			return pubsub.NackRequeue
		}

		return pubsub.Ack
	}
}
