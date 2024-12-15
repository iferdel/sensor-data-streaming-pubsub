package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
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

type AckType int

const (
	Ack AckType = iota
	NackDiscard
	NackRequeue
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueDurability QueueDurability,
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
		nil,                             // args
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
