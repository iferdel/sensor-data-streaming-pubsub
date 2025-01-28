package main

import (
	"context"
	"fmt"

	"github.com/iferdel/treanteyes/internal/pubsub"
	"github.com/iferdel/treanteyes/internal/routing"
	"github.com/iferdel/treanteyes/internal/sensorlogic"
	"github.com/iferdel/treanteyes/internal/storage"
)

func handlerMeasurements(db *storage.DB, ctx context.Context) func(m routing.SensorMeasurement) pubsub.AckType {
	return func(m routing.SensorMeasurement) pubsub.AckType {
		err := sensorlogic.HandleMeasurement(ctx, db, m)
		if err != nil {
			fmt.Printf("error writing sensor measurement instance: %v\n", err)
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
