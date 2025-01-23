package storage

import "time"

// TODO: add created_at and updated_at audit fields
type TargetRecord struct {
	ID   int
	Name string
}

type SensorRecord struct {
	ID              int     `json:"id"`
	SerialNumber    string  `json:"serial_number"`
	SampleFrequency float64 `json:"sample_frequency"`
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
