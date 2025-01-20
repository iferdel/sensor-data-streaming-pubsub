package sensorlogic

func (sensorState *SensorState) HandleAwake() {
	if sensorState.IsSleep {
		sensorState.IsSleep = false
		sensorState.LogsInfo <- "sensor is awake from sleep"
		return
	}
	sensorState.LogsInfo <- "sensor is already in an awake state"
}
