package main

import (
	"context"
	"fmt"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/sensorlogic"
	"github.com/iferdel/sensor-data-streaming-server/internal/storage"
)

func handlerMeasurements(ctx context.Context, db *storage.DB) func(m []routing.SensorMeasurement) pubsub.AckType {
	return func(m []routing.SensorMeasurement) pubsub.AckType {
		err := sensorlogic.HandleMeasurements(ctx, db, m)
		if err != nil {
			fmt.Printf("error writing sensor measurement instance: %v\n", err)
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}

func handlerMeasurementsWithCache(ctx context.Context, cache *sensorlogic.SensorCache, db *storage.DB) func(m []routing.SensorMeasurement) pubsub.AckType {
	return func(m []routing.SensorMeasurement) pubsub.AckType {
		start := time.Now()

		metricsMessagesReceived.Inc()
		metricsBatchSize.Observe(float64(len(m)))

		err := sensorlogic.HandleMeasurementsWithCache(ctx, cache, db, m)

		metricsProcessingDuration.Observe(time.Since(start).Seconds())

		for _, measurement := range m {
			metricsE2ELatency.Observe(time.Since(measurement.Timestamp).Seconds())
		}

		if err != nil {
			fmt.Printf("error writing sensor measurement instance: %v\n", err)
			metricsMeasurementsProcessed.WithLabelValues("error").Add(float64(len(m)))
			return pubsub.NackRequeue
		}

		metricsMeasurementsProcessed.WithLabelValues("success").Add(float64(len(m)))
		return pubsub.Ack
	}
}
