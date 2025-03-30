package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/storage"

	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
)

func main() {

	db, err := storage.NewDBPool(storage.PostgresConnString)
	if err != nil {
		msg := fmt.Sprintf("could not open pool connection to PostgreSQL: %v", err)
		fmt.Println(msg)
		return
	}
	defer db.Close()
	ctx := context.Background()

	env, err := stream.NewEnvironment(
		stream.NewEnvironmentOptions().SetUri(routing.RabbitStreamConnString),
	)
	if err != nil {
		fmt.Printf("Error creating stream environment: %v\n", err)
		return
	}

	// create stream
	streamName := "sensor.all.measurements.db_writer"
	err = env.DeclareStream(streamName, stream.NewStreamOptions().SetMaxLengthBytes(stream.ByteCapacity{}.GB(2)))
	if err != nil && !errors.Is(err, stream.StreamAlreadyExists) {
		fmt.Printf("Error declaring stream: %v\n", err)
		return
	}

	consumerName := "iot"
	consumerUpdate := func(streamName string, isActive bool) stream.OffsetSpecification {
		fmt.Printf("[%s] - Consumer promoted for: %s. Active status: %t\n", time.Now().Format(time.TimeOnly), streamName, isActive)
		offset, err := env.QueryOffset(consumerName, streamName)
		if err != nil {
			// If the offset is not found, we start from the beginning
			return stream.OffsetSpecification{}.First()
		}
		return stream.OffsetSpecification{}.Offset(offset + 1)
	}

	consumer, err := pubsub.SubscribeStreamJSON(
		env,
		streamName,
		stream.NewConsumerOptions().
			SetOffset(stream.OffsetSpecification{}.First()).
			SetConsumerName(consumerName).
			SetSingleActiveConsumer(stream.NewSingleActiveConsumer(consumerUpdate)),
		handlerMeasurements(db, ctx),
	)
	if err != nil {
		fmt.Println("error un subscribe stream json")
	}

	defer consumer.Close()
	defer env.Close()

	// Graceful shutdown handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Waiting for messages. Press Ctrl+C to exit.")
	<-sigs
	fmt.Println("Shutting down gracefully.")
}
