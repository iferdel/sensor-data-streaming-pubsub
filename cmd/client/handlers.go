package main

import (
	"fmt"

	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
)

func handlerSleep() func(routing.SensorStatus) {
	return func(sensorStatus routing.SensorStatus) {
		defer fmt.Println("a sleeeppppp")
	}
}
