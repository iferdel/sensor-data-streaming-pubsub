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
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		msg := fmt.Sprintf("could not connect to RabbitMQ: %v", err)
		fmt.Println(msg)
		return
	}
	defer conn.Close()

	// subscribe to Log queue
	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangeTopicIoT,
		routing.QueueIoTLogs,
		routing.BindKeyIoTLogs,
		pubsub.QueueDurable,
		handlerLogs(),
	)
	if err != nil {
		log.Fatalf("could not starting consuming logs: %v", err)
	}

	// Graceful shutdown handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Waiting for messages. Press Ctrl+C to exit.")
	<-sigs
	fmt.Println("Shutting down gracefully.")
}
