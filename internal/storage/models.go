package storage

import "time"

type SensorRecord struct {
	ID              int
	SerialNumber    string
	SampleFrequency float64
}

// timescaleDB hypertable -- Does not support primary keys
type SensorMeasurementRecord struct {
	Timestamp   time.Time
	SensorID    int
	Measurement float64
}

type SensorLogRecord struct {
	Timestamp    time.Time `json:"timestamp"`
	SerialNumber string    `json:"serialNumber"`
	Level        string    `json:"level"`
	Message      string    `json:"message"`
}
