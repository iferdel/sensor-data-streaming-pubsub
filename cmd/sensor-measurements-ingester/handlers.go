package main

import (
	"context"
	"fmt"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/sensorlogic"
	"github.com/iferdel/sensor-data-streaming-server/internal/storage"
)

func handlerMeasurements(db *storage.DB, ctx context.Context) func(m []routing.SensorMeasurement) pubsub.AckType {
	return func(m []routing.SensorMeasurement) pubsub.AckType {
		err := sensorlogic.HandleMeasurements(ctx, db, m)
		if err != nil {
			fmt.Printf("error writing sensor measurement instance: %v\n", err)
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
