package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/iferdel/treanteyes/internal/pubsub"
	"github.com/iferdel/treanteyes/internal/routing"
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

	// subscribe to Log queue
	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangeTopicIoT,
		routing.QueueSensorLogs,
		fmt.Sprintf(routing.KeySensorLogsFormat, "*")+"."+"#", // binding key
		pubsub.QueueDurable,
		handlerLogs(),
	)
	if err != nil {
		log.Fatalf("could not starting consuming logs: %v", err)
	}

	// Graceful shutdown handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Waiting for messages...")
	<-sigs
	fmt.Println("Shutting down gracefully...")
}
