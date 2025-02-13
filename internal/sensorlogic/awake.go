package sensorlogic

import "fmt"

func (sensorState *SensorState) HandleAwake() {
	// placeholder
	if sensorState.IsSleep {
		sensorState.IsSleep = false
		sensorState.IsSleepChan <- false
		sensorState.LogsInfo <- fmt.Sprintf("sensor is awake from sleep at %v [Hz]", sensorState.SampleFrequency)
		return
	}
	sensorState.LogsInfo <- "sensor is already in an awake state"
}
