package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Measurements processed counter
	MeasurementsProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sensor_measurements_processed_total",
			Help: "Total number of sensor measurements processed",
		},
		[]string{"status"}, // success, error
	)

	// Measurements per batch
	MeasurementsPerBatch = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "sensor_measurements_batch_size",
			Help:    "Number of measurements in each batch",
			Buckets: prometheus.ExponentialBuckets(1, 2, 20), // 1, 2, 4, 8, ... up to ~1M
		},
	)

	// Processing latency
	ProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sensor_measurements_processing_duration_seconds",
			Help:    "Time taken to process a batch of measurements",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to ~16s
		},
		[]string{"phase"}, // unmarshal, db_write, total
	)

	// Current throughput (measurements per second)
	CurrentThroughput = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "sensor_measurements_throughput_hz",
			Help: "Current throughput in measurements per second (Hz)",
		},
	)

	// Active sensors count
	ActiveSensorsCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "sensor_active_count",
			Help: "Number of active sensors sending data",
		},
	)

	// Database write errors
	DatabaseErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "sensor_measurements_db_errors_total",
			Help: "Total number of database write errors",
		},
	)

	// Messages received from RabbitMQ (before any processing)
	MessagesReceived = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "sensor_messages_received_total",
			Help: "Total messages received from RabbitMQ stream",
		},
	)

	// Deserialization/unmarshal errors (distinct from DB errors)
	UnmarshalErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "sensor_messages_unmarshal_errors_total",
			Help: "Total messages that failed to unmarshal",
		},
	)

	// End-to-end latency (from message creation to processing complete)
	EndToEndLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "sensor_message_e2e_latency_seconds",
			Help:    "Time from message creation to processing complete",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 12), // 10ms to ~40s
		},
	)
)
