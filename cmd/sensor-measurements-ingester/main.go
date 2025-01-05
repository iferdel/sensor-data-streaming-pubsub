package main

import (
	"fmt"
	"log"
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
	err = storage.CreateTableMeasurement()
	if err != nil {
		fmt.Printf("Error while creating/cheking `sensor_measurement` table: %v\n", err)
		return
	}

	// subscribe to Measurement queue
	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangeTopicIoT,
		routing.QueueSensorMeasurements,
		fmt.Sprintf(routing.KeySensorMeasurements, "*")+".#", // binding key
		pubsub.QueueDurable,
		handlerMeasurements(),
	)
	if err != nil {
		log.Fatalf("could not starting consuming measurements: %v", err)
	}

	// Graceful shutdown handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Waiting for messages. Press Ctrl+C to exit.")
	<-sigs
	fmt.Println("Shutting down gracefully.")
}
