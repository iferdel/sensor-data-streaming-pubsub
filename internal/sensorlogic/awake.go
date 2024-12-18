package sensorlogic

import "fmt"

func (sensorState *SensorState) HandleAwake() {
	if sensorState.IsSleep {
		sensorState.IsSleep = false
		fmt.Println("sensor is awake from sleep")
		return
	}
	fmt.Println("sensor is already in an awake state")
}
