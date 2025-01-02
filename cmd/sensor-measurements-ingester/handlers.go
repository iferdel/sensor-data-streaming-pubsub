package main

import (
	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/storage"
)

func handlerMeasurements() func(m routing.SensorMeasurement) pubsub.AckType {
	return func(m routing.SensorMeasurement) pubsub.AckType {
		storage.WriteMeasurement(m)
		return pubsub.Ack
	}
}
