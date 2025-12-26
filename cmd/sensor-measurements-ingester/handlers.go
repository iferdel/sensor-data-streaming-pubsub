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

		// Count measurements and track oldest timestamp per sensor (single pass)
		// Using oldest timestamp gives worst-case E2E latency per batch
		type sensorStats struct {
			count           int
			oldestTimestamp time.Time
		}
		sensorData := make(map[string]*sensorStats)
		for _, measurement := range m {
			stats, exists := sensorData[measurement.SerialNumber]
			if !exists {
				sensorData[measurement.SerialNumber] = &sensorStats{
					count:           1,
					oldestTimestamp: measurement.Timestamp,
				}
			} else {
				stats.count++
				if measurement.Timestamp.Before(stats.oldestTimestamp) {
					stats.oldestTimestamp = measurement.Timestamp
				}
			}
		}

		err := sensorlogic.HandleMeasurementsWithCache(ctx, cache, db, m)

		metricsProcessingDuration.Observe(time.Since(start).Seconds())

		// Record E2E latency once per sensor per batch (using oldest = worst case latency)
		processingComplete := time.Now()
		for serial, stats := range sensorData {
			metricsE2ELatency.WithLabelValues(serial).Observe(processingComplete.Sub(stats.oldestTimestamp).Seconds())
		}

		if err != nil {
			fmt.Printf("error writing sensor measurement instance: %v\n", err)
			for serial, stats := range sensorData {
				metricsMeasurementsProcessed.WithLabelValues("error", serial).Add(float64(stats.count))
			}
			return pubsub.NackRequeue
		}

		for serial, stats := range sensorData {
			metricsMeasurementsProcessed.WithLabelValues("success", serial).Add(float64(stats.count))
		}
		return pubsub.Ack
	}
}
