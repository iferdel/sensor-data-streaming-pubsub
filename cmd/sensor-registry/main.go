package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	// "github.com/iferdel/sensor-data-streaming-server/internal/storage"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	conn, err := amqp.Dial(routing.RabbitConnString)
	if err != nil {
		msg := fmt.Sprintf("could not connect to RabbitMQ: %v", err)
		fmt.Println(msg)
		return
	}
	defer conn.Close()

	// create sensor table (ignored if already created)
	// storage.CreateTableSensor()

	// consume sensor registration
	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangeTopicIoT,
		routing.QueueSensorTelemetryFormat,
		routing.BindKeySensorDataFormat,
		pubsub.QueueDurable,
		handlerSensorRegistry(), // consumption
	)
	if err != nil {
		fmt.Println("Could not subscribe to registry:", err)
		return
	}

	// publish trigger for sensor to start telemetry

	// Graceful shutdown handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Waiting for messages. Press Ctrl+C to exit.")
	<-sigs
	fmt.Println("Shutting down gracefully.")
}
