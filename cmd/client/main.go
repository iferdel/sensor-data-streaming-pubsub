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
	go sensorOperation(&wg, "AAD-1123", 1*time.Second, 99)
	wg.Wait() // it blocks the execution of whatever comes next until all goroutines it's waiting are finished
}

// each sensor as a client that would run in a different process or all sensors as a client (for simplicity)
func sensorOperation(wg *sync.WaitGroup, serialNumber string, interval time.Duration, seed int64) {
	defer wg.Done() // signals the waitGroup that the goroutine finished its job, bringing the counter down a unit value
	fmt.Println("EQP ON")

	const rabbitConnString = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}

	defer conn.Close()
	fmt.Println("connection to msg broker succeeded")

	sensorState := sensorlogic.NewSensorState(serialNumber)

	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.QueueSensorCommandsFormat, serialNumber),  // queue name
		fmt.Sprintf(routing.BindKeySensorCommandFormat, serialNumber), // routing key
		pubsub.QueueDurable, // queue type
		handlerCommand(sensorState),
	)
	if err != nil {
		log.Fatalf("could not subscribe to sleep: %v", err)
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
