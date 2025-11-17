package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/sensorlogic"
	"github.com/iferdel/sensor-data-streaming-server/internal/storage"

	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := storage.NewDBPool(storage.PostgresConnString)
	if err != nil {
		msg := fmt.Sprintf("could not open pool connection to PostgreSQL: %v", err)
		fmt.Println(msg)
		return
	}
	defer db.Close()

	sensorCache, err := sensorlogic.NewSensorCache(ctx, db)
	if err != nil {
		fmt.Printf("Failed to initialize sensor cache: %v\n", err)
		return
	}

	go sensorCache.StartRefreshLoop(ctx)

	env, err := stream.NewEnvironment(
		stream.NewEnvironmentOptions().SetUri(routing.RabbitStreamConnString),
	)
	if err != nil {
		fmt.Printf("Error creating stream environment: %v\n", err)
		return
	}
	defer env.Close()

	singleActiveConsumerUpdate := func(streamName string, isActive bool) stream.OffsetSpecification {
		fmt.Printf("[%s] - Consumer promoted for: %s. Active status: %t\n", time.Now().Format(time.TimeOnly), streamName, isActive)
		offset, err := env.QueryOffset(routing.StreamConsumerName, routing.QueueSensorMeasurements)
		if err != nil {
			// If the offset is not found, we start from the beginning
			return stream.OffsetSpecification{}.First()
		}
		return stream.OffsetSpecification{}.Offset(offset + 1)
	}

	consumer, err := pubsub.SubscribeStreamJSON(
		env,
		routing.QueueSensorMeasurements,
		stream.NewConsumerOptions().
			SetOffset(stream.OffsetSpecification{}.First()).
			SetConsumerName(routing.StreamConsumerName).
			SetSingleActiveConsumer(stream.NewSingleActiveConsumer(singleActiveConsumerUpdate)),
		handlerMeasurements(db, ctx),
		// handlerMeasurementsWithCache(sensorCache, db, ctx),
	)
	if err != nil {
		fmt.Println("error un subscribe stream json")
	}

	defer consumer.Close()

	// Graceful shutdown handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Waiting for messages. Press Ctrl+C to exit.")
	<-sigs
	fmt.Println("Shutting down gracefully...")
	cancel()                           // This stops the refresh loop
	time.Sleep(100 * time.Millisecond) // Give goroutines time to clean up
	fmt.Println("Shutdown complete.")
}
