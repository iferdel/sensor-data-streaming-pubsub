package main

import (
	"fmt"
	"log"

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

	pubsub.PublishGob(
		publishCh,                               // amqp.Channel
		routing.ExchangeSensorTransmissionTopic, // exchange
		routing.PauseKey+".*",                   // routing key
		routing.SensorState{
			IsPaused: true,
		}, // value
	)
}
