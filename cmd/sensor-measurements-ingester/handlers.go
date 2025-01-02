package main

import (
	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
)

func handlerMeasurements() func() pubsub.AckType {
	return func() pubsub.AckType {
		return pubsub.Ack
	}
}
