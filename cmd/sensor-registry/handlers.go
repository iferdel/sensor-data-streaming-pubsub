package main

import (
	"fmt"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/storage"
)

func handlerSensorRegistry() func(dto routing.Sensor) pubsub.AckType {
	return func(dto routing.Sensor) pubsub.AckType {
		// placeholder
		fmt.Println("==========================================")

		// Map DTO -to- DB Record
		record := storage.SensorRecord{
			SerialNumber:    dto.SerialNumber,
			SampleFrequency: dto.SampleFrequency,
		}

		err := storage.WriteSensor(record)
		if err != nil {
			fmt.Printf("error writing sensor instance: %v\n", err)
			return pubsub.NackRequeue
		}

		return pubsub.Ack
	}
}
