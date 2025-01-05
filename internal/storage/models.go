package storage

import "time"

type SensorRecord struct {
	ID           int
	SerialNumber string
}

// timescaleDB hypertable -- Does not support primary keys
type SensorMeasurementRecord struct {
	Time        time.Time
	SensorID    int
	Measurement float64
}

type SensorLogRecord struct {
	Timestamp    time.Time
	SerialNumber string
	Level        string
	Message      string
}
