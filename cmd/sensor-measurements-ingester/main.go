package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/iferdel/treanteyes/internal/pubsub"
	"github.com/iferdel/treanteyes/internal/routing"
	"github.com/iferdel/treanteyes/internal/storage"
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

	db, err := storage.NewDBPool(storage.PostgresConnString)
	defer db.Close()
	ctx := context.Background()

	// subscribe to Measurement queue
	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangeTopicIoT,
		routing.QueueSensorMeasurements,
		fmt.Sprintf(routing.KeySensorMeasurements, "*")+".#", // binding key
		pubsub.QueueDurable,
		handlerMeasurements(db, ctx),
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
