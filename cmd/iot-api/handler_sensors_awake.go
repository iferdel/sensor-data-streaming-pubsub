package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/iferdel/treanteyes/internal/pubsub"
	"github.com/iferdel/treanteyes/internal/routing"
)

func (cfg *apiConfig) handlerSensorsAwake(w http.ResponseWriter, req *http.Request) {
	sensorSerialNumber := req.PathValue("sensorSerialNumber")

	publishCh, err := cfg.rabbitConn.Channel()
	defer publishCh.Close()
	if err != nil {
		respondWithError(w, 500, "could not create channel to publish sensor's awake command", err)
		return
	}
	err = pubsub.PublishGob(
		publishCh,                // amqp.Channel
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.KeySensorCommandsFormat, sensorSerialNumber)+"."+"awake", // routing key
		routing.SensorCommandMessage{
			SerialNumber: sensorSerialNumber,
			Timestamp:    time.Now(),
			Command:      "awake",
			Params:       nil,
		}, // value
	)
	if err != nil {
		log.Printf("could not publish awake command: %v", err)
	}

}
