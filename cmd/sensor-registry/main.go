package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/storage"

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
	err = storage.CreateTableSensor()
	if err != nil {
		fmt.Printf("Error while creating/cheking sensor table: %v\n", err)
		return
	}

	// consume sensor registration
	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangeTopicIoT,
		routing.QueueSensorRegistry,
		fmt.Sprintf(routing.KeySensorRegistryFormat, "*")+"."+"#", // subscribeGob creates and bind a queue to an exchange in case it is not yet there. Thats why here we have binding key (and not just queue name)
		pubsub.QueueDurable,
		handlerSensorRegistry(), // consumption
	)
	if err != nil {
		fmt.Println("Could not subscribe to registry:", err)
		return
	}

	// publish trigger for sensor to start telemetry
	// the broker can confirm the producer that the msg was received

	// Graceful shutdown handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Waiting for messages. Press Ctrl+C to exit.")
	<-sigs
	fmt.Println("Shutting down gracefully.")
}
