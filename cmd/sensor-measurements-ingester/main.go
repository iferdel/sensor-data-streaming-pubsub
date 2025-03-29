package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/storage"

	amqp "github.com/rabbitmq/amqp091-go"

	amqpForStream "github.com/rabbitmq/rabbitmq-stream-go-client/pkg/amqp"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/ha"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
)

func main() {

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

	// possibly, when declaring singe active consumer maybe the stream queue is not needed to be created yet BUT
	// it does not make sense since this is altering the consumers and not the queues. The declaration of the stream/queue is in
	// the DeclareStream method
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
	consumer, err := ha.NewReliableConsumer(
		env,
		streamName,
		// start from the beginning of the stream
		stream.NewConsumerOptions().
			SetOffset(stream.OffsetSpecification{}.First()).
			SetConsumerName(consumerName).
			SetSingleActiveConsumer(stream.NewSingleActiveConsumer(consumerUpdate)),
		func(consumerContext stream.ConsumerContext, message *amqpForStream.Message) {
			fmt.Printf("Message received: %s\n", message.GetData())
		},
	)
	if err != nil {
		fmt.Printf("Error creating stream consumer: %v\n", err)
	}
	defer consumer.Close()
	defer env.Close()

	conn, err := amqp.Dial(routing.RabbitConnString)
	if err != nil {
		msg := fmt.Sprintf("could not connect to RabbitMQ: %v", err)
		fmt.Println(msg)
		return
	}
	defer conn.Close()

	db, err := storage.NewDBPool(storage.PostgresConnString)
	if err != nil {
		msg := fmt.Sprintf("could not open pool connection to PostgreSQL: %v", err)
		fmt.Println(msg)
		return
	}
	defer db.Close()
	_ = context.Background()

	// // subscribe to Measurement queue
	// err = pubsub.SubscribeJSON(
	// 	conn,
	// 	routing.ExchangeTopicIoT,
	// 	routing.QueueSensorMeasurements,
	// 	fmt.Sprintf(routing.KeySensorMeasurements, "*")+".#", // binding key
	// 	pubsub.QueueDurable,
	// 	pubsub.QueueStream,
	// 	handlerMeasurements(db, ctx),
	// )
	// if err != nil {
	// 	log.Fatalf("could not starting consuming measurements: %v", err)
	// }

	// Graceful shutdown handling
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Waiting for messages. Press Ctrl+C to exit.")
	<-sigs
	fmt.Println("Shutting down gracefully.")
}
