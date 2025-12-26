package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// metricsMessagesReceived counts messages received from RabbitMQ (before any processing)
	metricsMessagesReceived = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "sensor_measurements_messages_received_total",
			Help: "Total messages received from RabbitMQ stream",
		},
	)

	// metricsMeasurementsProcessed counts measurements by processing status and sensor
	metricsMeasurementsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sensor_measurements_processed_total",
			Help: "Total number of sensor measurements processed",
		},
		[]string{"status", "sensor_serial"}, // status: success/error, sensor_serial: sensor identifier
	)

	// metricsBatchSize tracks the distribution of batch sizes
	metricsBatchSize = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "sensor_measurements_batch_size",
			Help:    "Number of measurements in each batch",
			Buckets: prometheus.ExponentialBuckets(1, 2, 20), // 1, 2, 4, 8, ... up to ~1M
		},
	)

	// metricsProcessingDuration tracks time spent processing a batch
	metricsProcessingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "sensor_measurements_processing_duration_seconds",
			Help:    "Time taken to process a batch of measurements",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to ~16s
		},
	)

	// metricsE2ELatency tracks time from measurement creation to processing complete per sensor
	metricsE2ELatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sensor_measurements_e2e_latency_seconds",
			Help:    "Time from measurement creation to processing complete",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 12), // 10ms to ~40s
		},
		[]string{"sensor_serial"},
	)
)
