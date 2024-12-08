package routing

import "time"

type SensorState struct {
    IsPaused bool
}

type GameLog struct {
    CurrentTime time.Time
    Message string
    SensorName string
}
