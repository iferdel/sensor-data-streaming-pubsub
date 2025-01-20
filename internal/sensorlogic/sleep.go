package sensorlogic

func (sensorState *SensorState) HandleSleep() {
	if sensorState.IsSleep {
		sensorState.LogsInfo <- "sensor is already in a sleep state"
		return
	}
	sensorState.IsSleep = true
	sensorState.LogsInfo <- "sensor is set to sleep"
}
