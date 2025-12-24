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
	rabbitConn *amqp.Connection
	db         *storage.DB
}

func NewApiConfig() (*apiConfig, error) {
	conn, err := amqp.Dial(routing.RabbitConnString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	db, err := storage.NewDBPool(storage.PostgresConnString)

	return &apiConfig{
		rabbitConn: conn,
		db:         db,
	}, nil
}

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	apiCfg, err := NewApiConfig()
	if err != nil {
		log.Fatal(err)
	}
	defer apiCfg.rabbitConn.Close()
	defer apiCfg.db.Close()

	// consume sensor registration
	err = pubsub.SubscribeGob(
		ctx,
		apiCfg.rabbitConn,
		routing.ExchangeTopicIoT,
		routing.QueueSensorRegistry,
		fmt.Sprintf(routing.KeySensorRegistryFormat, "*")+"."+"#", // subscribeGob creates and bind a queue to an exchange in case it is not yet there. Thats why here we have binding key (and not just queue name)
		pubsub.QueueDurable,
		pubsub.QueueClassic,
		handlerSensorRegistry(ctx, apiCfg.db), // consumption
	)
	if err != nil {
		fmt.Println("Could not subscribe to registry:", err)
		return
	}

	// publish trigger for sensor to start telemetry
	// the broker can confirm the producer that the msg was received

	// Graceful shutdown handling
	fmt.Println("Waiting for messages. Press Ctrl+C to exit.")
	<-ctx.Done()
	fmt.Println("Shutting down gracefully.")
}
