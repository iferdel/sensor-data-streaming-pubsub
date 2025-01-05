package storage

import "time"

type measurement struct {
	Time        time.Time
	SensorId    int
	Measurement float64
}
