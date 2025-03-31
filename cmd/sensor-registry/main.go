package main

import (
	"context"
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

type apiConfig struct {
	ctx        context.Context
	rabbitConn *amqp.Connection
	db         *storage.DB
}

func NewApiConfig() (*apiConfig, error) {
	ctx := context.Background()

	conn, err := amqp.Dial(routing.RabbitConnString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	db, err := storage.NewDBPool(storage.PostgresConnString)

	return &apiConfig{
		ctx:        ctx,
		rabbitConn: conn,
		db:         db,
	}, nil
}

func main() {

	apiCfg, err := NewApiConfig()
	if err != nil {
		log.Fatal(err)
	}
	defer apiCfg.rabbitConn.Close()
	defer apiCfg.db.Close()

	// consume sensor registration
	err = pubsub.SubscribeGob(
		apiCfg.rabbitConn,
		routing.ExchangeTopicIoT,
		routing.QueueSensorRegistry,
		fmt.Sprintf(routing.KeySensorRegistryFormat, "*")+"."+"#", // subscribeGob creates and bind a queue to an exchange in case it is not yet there. Thats why here we have binding key (and not just queue name)
		pubsub.QueueDurable,
		pubsub.QueueClassic,
		handlerSensorRegistry(apiCfg.ctx, apiCfg.db), // consumption
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
