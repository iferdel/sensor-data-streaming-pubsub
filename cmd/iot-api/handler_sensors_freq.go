package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (cfg *apiConfig) handlerSensorsChangeSampleFrequency(w http.ResponseWriter, req *http.Request) {
	// TODO: relation with database, how to keep state between sensor current sample frequency and db registered sample frequency.
	// Maybe this last point (saving sample frequency in db) is redundant and useless

	defer req.Body.Close()

	sensorSerialNumber := req.PathValue("sensorSerialNumber")

	decoder := json.NewDecoder(req.Body)
	type parameters struct {
		NewSampleFrequency float64 `json:"new_sample_frequency"`
	}
	params := parameters{}
	decoder.Decode(&params)

	publishCh, err := cfg.rabbitConn.Channel()
	defer publishCh.Close()
	if err != nil {
		respondWithError(w, 500, "could not create channel to publish sensor's new sample frequency:", err)
		return
	}
	err = pubsub.PublishGob(
		publishCh,                // amqp.Channel
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.KeySensorCommandsFormat, sensorSerialNumber)+"."+"change_sample_frequency", // routing key
		routing.SensorCommandMessage{
			SerialNumber: sensorSerialNumber,
			Timestamp:    time.Now(),
			Command:      "changeSampleFrequency",
			Params: map[string]interface{}{
				"sampleFrequency": params.NewSampleFrequency,
			},
		}, // value
	)
	// publish sensor logs
	err = publishSensorLog(
		publishCh,
		routing.SensorLog{
			SerialNumber: sensorSerialNumber,
			Timestamp:    time.Now(),
			Level:        "INFO",
			Message:      fmt.Sprintf("Sample frequency changed to %v [Hz]", params.NewSampleFrequency),
		},
	)
	if err != nil {
		fmt.Printf("Error publishing log: %s\n", err)
	}
	if err != nil {
		log.Printf("could not publish change sample frequency command: %v", err)
	}
}

func publishSensorLog(publishCh *amqp.Channel, sensorLog routing.SensorLog) error {
	return pubsub.PublishGob(
		publishCh,                // channel
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.KeySensorLogsFormat, sensorLog.SerialNumber)+"."+"boot", // routing key
		sensorLog, // sensor log
	)
}
