package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conn, err := amqp.Dial(routing.RabbitConnString)
	if err != nil {
		msg := fmt.Sprintf("could not connect to RabbitMQ: %v", err)
		fmt.Println(msg)
		return
	}
	defer conn.Close()

	// subscribe to Log queue
	err = pubsub.SubscribeGob(
		ctx,
		conn,
		routing.ExchangeTopicIoT,
		routing.QueueSensorLogs,
		fmt.Sprintf(routing.KeySensorLogsFormat, "*")+"."+"#", // binding key
		pubsub.QueueDurable,
		pubsub.QueueQuorum,
		handlerLogs(),
	)
	if err != nil {
		log.Fatalf("could not starting consuming logs: %v", err)
	}

	// Graceful shutdown handling
	fmt.Println("Waiting for messages...")
	<-ctx.Done()
	fmt.Println("Shutting down gracefully...")
}
