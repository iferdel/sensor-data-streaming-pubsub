package sensorlogic

import (
	"fmt"
)

func (sensorState *SensorState) HandleChangeSampleFrequency(params map[string]interface{}) {
	if sampleFrequency, ok := params["sampleFrequency"]; ok {
		sensorState.SampleFrequency = sampleFrequency.(float64)
		// signal the channel of the change of sample frequency
		sensorState.SampleFrequencyChangeChan <- sampleFrequency.(float64)
		if sensorState.IsSleep {
			fmt.Println("changes of sample frequency applied, but sensor is currently in a sleep state")
			return
		}
		fmt.Println("changes of sample frequency applied")
	} else {
		fmt.Println("SampleFrequency is not a number")
	}
}
