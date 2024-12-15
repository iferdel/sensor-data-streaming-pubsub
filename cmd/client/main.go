package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/sensorlogic"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(2) // Increment the wait count by 2, since we will have 2 goroutines calling Done(). It counts at zero will trigger Wait() and unblock the program.
	go sensorOutput(&wg, "AAD-1123", 1*time.Second, 99)
	wg.Wait() // it blocks the execution of whatever comes next until all goroutines it's waiting are finished
}

func sensorOutput(wg *sync.WaitGroup, serialNumber string, interval time.Duration, seed int64) {
	defer wg.Done() // signals the waitGroup that the goroutine finished its job, bringing the counter down a unit value
	fmt.Println("EQP ON")

	const rabbitConnString = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}

	defer conn.Close()
	fmt.Println("connection to msg broker succeeded")

	// create channel for further publish of sensor data/logs
	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create publish channel: %v", err)
	}

	_ = sensorlogic.NewSensorState(serialNumber)

	err = publishSensorLog(
		publishCh,
		serialNumber,
		"Starting Sensor Streaming...",
	)
	if err != nil {
		fmt.Println("invalid publish sensor log:", err)
	}

	// consumer of command queue
	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.QueueSensorCommandsFormat, serialNumber), // queue name
		fmt.Sprintf(routing.KeySensorCommandFormat, serialNumber),    // routing key
		pubsub.SimpleQueueDurable,                                    // queue type
	)
	if err != nil {
		log.Fatalf(
			"error declaring and binding on exchange %v, queue %v, routing key %v: %v",
			routing.ExchangeTopicIoT,                                     // exchange
			fmt.Sprintf(routing.QueueSensorCommandsFormat, serialNumber), // queue name
			fmt.Sprintf(routing.KeySensorCommandFormat, serialNumber),    // routing key
			err,
		)
	}

	// publisher of data streaming queue
	_, _, err = pubsub.DeclareAndBind(
		conn,
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.QueueSensorTelemetryFormat, serialNumber), // queue name
		fmt.Sprintf(routing.KeySensorDataFormat, serialNumber),        // routing key
		pubsub.SimpleQueueDurable,                                     // queue type
	)
	if err != nil {
		log.Fatalf(
			"error declaring and binding on exchange %v, queue %v, routing key %v: %v",
			routing.ExchangeTopicIoT, // exchange
			fmt.Sprintf(routing.QueueSensorTelemetryFormat, serialNumber), // queue name
			fmt.Sprintf(routing.KeySensorDataFormat, serialNumber),        // routing key
			err,
		)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop() // stop Ticker on return so no more ticks will be sent and thus freeing resources

	r := rand.New(rand.NewSource(seed))
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)

	show := func(name string, accX, accY, accZ any) {
		fmt.Fprintf(w, "%s\t%v\t%v\t%v\n", name, accX, accY, accZ)
	}
	for range ticker.C {
		show(serialNumber, r.Float64(), r.Float64(), r.Float64())
		w.Flush() // allows to write buffered output from tabwriter to stdout immediatly
	}
}

func publishSensorLog(publishCh *amqp.Channel, sensorname, msg string) error {
	return pubsub.PublishGob(
		publishCh,
		routing.ExchangeTopicIoT,
		routing.SensorLogSlug+"."+sensorname,
		routing.SensorLog{
			SensorName:  sensorname,
			CurrentTime: time.Now(),
			Message:     msg,
		},
	)
}
