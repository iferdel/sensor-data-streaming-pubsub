package routing

import "time"

type SensorState struct {
	IsPaused bool
}

type SensorLog struct {
	SensorName  string
	CurrentTime time.Time
	Message     string
}
