package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
)

func (cfg *apiConfig) handlerSensorsSleep(w http.ResponseWriter, req *http.Request) {
	sensorSerialNumber := req.PathValue("sensorSerialNumber")

	publishCh, err := cfg.rabbitConn.Channel()
	if err != nil {
		respondWithError(w, 500, "could not create channel to publish sensor's sleep command", err)
		return
	}
	defer publishCh.Close()

	err = pubsub.PublishGob(
		publishCh,                // amqp.Channel
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.KeySensorCommandsFormat, sensorSerialNumber)+"."+"sleep", // routing key
		routing.SensorCommandMessage{
			SerialNumber: sensorSerialNumber,
			Timestamp:    time.Now(),
			Command:      "sleep",
			Params:       nil,
		}, // value
	)
	if err != nil {
		log.Printf("could not publish sleep command: %v", err)
	}

}
