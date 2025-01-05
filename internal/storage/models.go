package storage

import "time"

type SensorRecord struct {
	ID           int
	SerialNumber string
}

// timescaleDB hypertable -- Does not support primary keys
type MeasurementRecord struct {
	Time        time.Time
	SensorID    int
	Measurement float64
}
