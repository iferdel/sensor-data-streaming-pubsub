package sensorlogic

import (
	"fmt"
)

func (sensorState *SensorState) HandleChangeSampleFrequency(params map[string]interface{}) {
	if sampleFrequency, ok := params["sampleFrequency"]; ok {
		sensorState.SampleFrequency = sampleFrequency.(int)
		// signal the channel of the change of sample frequency
		sensorState.SampleFrequencyChangeChan <- sampleFrequency.(int)
		if sensorState.IsSleep {
			fmt.Println("changes of sample frequency applied, but sensor is currently in a sleep state")
			return
		}
		fmt.Println("changes of sample frequency applied")
	} else {
		fmt.Println("SampleFrequency is not an integer")
	}
}
