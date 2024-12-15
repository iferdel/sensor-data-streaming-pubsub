package main

import (
	"fmt"
	"log"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	const rabbitConnString = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Server connected to RabbitMQ")

	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	sensorSerialNumber := "AAD-1123"

	pubsub.PublishGob(
		publishCh,                // amqp.Channel
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.BindKeySensorCommandFormat, sensorSerialNumber), // routing key
		routing.CommandMessage{
			SensorName: sensorSerialNumber,
			Timestamp:  time.Now(),
			Command:    "sleep",
			Params:     nil,
		}, // value
	)
}
