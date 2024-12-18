package sensorlogic

import "fmt"

func (sensorState *SensorState) HandleSleep() {
	if sensorState.IsSleep {
		fmt.Println("sensor is already in a sleep state")
		return
	}
	sensorState.IsSleep = true
	fmt.Println("sensor is set to sleep")
}
