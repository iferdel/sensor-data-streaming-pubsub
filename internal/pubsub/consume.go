package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	amqpForStream "github.com/rabbitmq/rabbitmq-stream-go-client/pkg/amqp"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/ha"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
)

type QueueDurability int

type QueueType int

const (
	QueueDurable QueueDurability = iota
	QueueTranscient
)

const (
	QueueClassic QueueType = iota
	QueueQuorum
	QueueStream
)

func (qt QueueType) String() string {
	return [...]string{"classic", "quorum", "stream"}[qt]
}

type AckType int

const (
	Ack AckType = iota
	NackDiscard
	NackRequeue
)

func SubscribeStreamJSON[T any](
	env *stream.Environment,
	streamName string,
	streamOptions *stream.ConsumerOptions,
	handler func(T) AckType,
) (*ha.ReliableConsumer, error) {
	fmt.Println("subscribing to stream json")
	consumer, err := ha.NewReliableConsumer(
		env,
		streamName,
		streamOptions,
		func(consumerContext stream.ConsumerContext, message *amqpForStream.Message) {
			fmt.Println("unmarshalling stream data")
			var target T
			err := json.Unmarshal(message.GetData(), &target)
			if err != nil {
				fmt.Printf("could not unmarshal message: %v\n", err)
			}
			handler(target)
		},
	)
	return consumer, err
}

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueDurability QueueDurability,
	queueType QueueType,
	handler func(T) AckType,
) error {
	return subscribe[T](
		conn,
		exchange,
		queueName,
		key,
		queueDurability,
		queueType,
		handler,
		func(data []byte) (T, error) {
			var target T
			err := json.Unmarshal(data, &target)
			return target, err
		},
	)
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueDurability QueueDurability,
	queueType QueueType,
	handler func(T) AckType,
) error {
	return subscribe[T](
		conn,
		exchange,
		queueName,
		key,
		queueDurability,
		queueType,
		handler,
		func(data []byte) (T, error) {
			buffer := bytes.NewBuffer(data)
			decoder := gob.NewDecoder(buffer)
			var target T
			err := decoder.Decode(&target)
			return target, err
		},
	)

}

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueDurability QueueDurability,
	queueType QueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
) error {
	ch, queue, err := DeclareAndBind(
		conn,
		exchange,
		queueName,
		key,
		queueDurability,
		queueType,
	)
	if err != nil {
		return fmt.Errorf("could not declare and bind queue: %v", err)
	}

	err = ch.Qos(10, 0, false) // luckily enough stream queues does not support global QoS prefetch
	if err != nil {
		return fmt.Errorf("could not set QoS: %v", err)
	}

	msgs, err := ch.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return fmt.Errorf("could not consume messages: %v", err)
	}

	go func() {
		defer ch.Close()
		for msg := range msgs {
			target, err := unmarshaller(msg.Body)
			if err != nil {
				fmt.Printf("could not unmarshal message: %v\n", err)
				continue
			}
			handler(target)
			msg.Ack(false)
		}
	}()

	return nil

}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueDurability QueueDurability,
	queueType QueueType,
) (*amqp.Channel, amqp.Queue, error) {

	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("error while opening channel in declare and bind: %v", err)
	}

	queue, err := ch.QueueDeclare(
		queueName,                       // name
		queueDurability == QueueDurable, // durable
		queueDurability != QueueDurable, // delete when unused
		queueDurability != QueueDurable, // exclusive
		false,                           // noWait
		amqp.Table{
			"x-queue-type": queueType.String(),
		}, // args
	)

	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not declare queue: %v", err)
	}

	err = ch.QueueBind(
		queue.Name, // name
		key,        // routing key
		exchange,   // exchange
		false,      // noWait
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not bind queue: %v", err)
	}

	return ch, queue, nil
}
