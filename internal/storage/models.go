package storage

import "time"

// TODO: add created_at and updated_at audit fields
type TargetRecord struct {
	ID   int
	Name string
}

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
