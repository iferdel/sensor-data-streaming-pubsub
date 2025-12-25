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

	// metricsMeasurementsProcessed counts measurements by processing status
	metricsMeasurementsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sensor_measurements_processed_total",
			Help: "Total number of sensor measurements processed",
		},
		[]string{"status"}, // success, error
	)

	// metricsBatchSize tracks the distribution of batch sizes
	metricsBatchSize = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "sensor_measurements_batch_size",
			Help:    "Number of measurements in each batch",
			Buckets: prometheus.ExponentialBuckets(1, 2, 20), // 1, 2, 4, 8, ... up to ~1M
		},
	)

	// metricsProcessingDuration tracks time spent in each processing phase
	metricsProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sensor_measurements_processing_duration_seconds",
			Help:    "Time taken to process a batch of measurements",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to ~16s
		},
		[]string{"phase"}, // unmarshal, db_write, total
	)

	// metricsThroughput tracks current measurements per second
	metricsThroughput = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "sensor_measurements_throughput_hz",
			Help: "Current throughput in measurements per second (Hz)",
		},
	)

	// metricsActiveSensors tracks number of sensors currently sending data
	metricsActiveSensors = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "sensor_measurements_active_sensors",
			Help: "Number of active sensors sending data",
		},
	)

	// metricsDbErrors counts database write failures
	metricsDbErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "sensor_measurements_db_errors_total",
			Help: "Total number of database write errors",
		},
	)

	// metricsUnmarshalErrors counts deserialization failures
	metricsUnmarshalErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "sensor_measurements_unmarshal_errors_total",
			Help: "Total messages that failed to unmarshal",
		},
	)

	// metricsE2ELatency tracks time from message creation to processing complete
	metricsE2ELatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "sensor_measurements_e2e_latency_seconds",
			Help:    "Time from message creation to processing complete",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 12), // 10ms to ~40s
		},
	)
)
