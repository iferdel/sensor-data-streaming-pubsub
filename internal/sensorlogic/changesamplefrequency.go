package sensorlogic

import (
	"fmt"
)

func (sensorState *SensorState) HandleChangeSampleFrequency(params map[string]interface{}) {
	var sampleFrequency float64

	if sf, ok := params["sampleFrequency"]; ok {
		sampleFrequency = sf.(float64)

		sensorState.SampleFrequency = sampleFrequency
		// signal the channel of the change of sample frequency
		sensorState.SampleFrequencyChangeChan <- sampleFrequency
		sensorState.LogsInfo <- fmt.Sprintf("Sample frequency changed to %v [Hz]", sampleFrequency)

		if sensorState.IsSleep {
			sensorState.LogsInfo <- "changes of sample frequency applied, but sensor is currently in a sleep state"
		}
		fmt.Println("changes of sample frequency applied")
	} else {
		sensorState.LogsWarning <- "SampleFrequency is not a number. Skipping..."
	}
}
