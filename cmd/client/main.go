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
	go sensorOperation(&wg, "AAD-1123", 1, 99)
	wg.Wait() // it blocks the execution of whatever comes next until all goroutines it's waiting are finished
}

// each sensor as a client that would run in a different process or all sensors as a client (for simplicity)
func sensorOperation(wg *sync.WaitGroup, serialNumber string, sampleFrequency int, seed int64) {
	defer wg.Done() // signals the waitGroup that the goroutine finished its job, bringing the counter down a unit value

	bootLogs := []string{}

	bootLogs = append(bootLogs, "EQP ON")

	const rabbitConnString = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		msg := fmt.Sprintf("could not connect to RabbitMQ: %v", err)
		bootLogs = append(bootLogs, msg)
		log.Fatalf(msg)
	}

	defer conn.Close()

	bootLogs = append(bootLogs, "connection to msg broker succeeded")

	sensorState := sensorlogic.NewSensorState(serialNumber, sampleFrequency)

	// subscribe to sensor command queue
	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.QueueSensorCommandsFormat, serialNumber),  // queue name
		fmt.Sprintf(routing.BindKeySensorCommandFormat, serialNumber), // routing key
		pubsub.QueueDurable, // queue type
		handlerCommand(sensorState),
	)
	if err != nil {
		msg := fmt.Sprintf("could not subscribe to command: %v", err)
		bootLogs = append(bootLogs, msg)
		log.Fatalf(msg)
	}

	publishCh, err := conn.Channel()
	if err != nil {
		msg := fmt.Sprintf("could not create channel: %v", err)
		bootLogs = append(bootLogs, msg)
		log.Fatalf(msg)
	}

	// no guarda la hora del comando, sólo de cuando se envían todos los boot logs
	for _, bootLog := range bootLogs {
		err = publishSensorLog(
			publishCh,
			sensorState.Sensor.SerialNumber,
			bootLog,
		)
		if err != nil {
			fmt.Printf("error publishing log: %s\n", err)
		}
	}

	ticker := time.NewTicker(time.Second / time.Duration(sensorState.SampleFrequency))
	defer ticker.Stop() // stop Ticker on return so no more ticks will be sent and thus freeing resources

	r := rand.New(rand.NewSource(seed))
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)

	show := func(name string, accX, accY, accZ any) {
		fmt.Fprintf(w, "%s\t%v\t%v\t%v\n", name, accX, accY, accZ)
	}
	for {
		select {
		case <-ticker.C:
			show(serialNumber, r.Float64(), r.Float64(), r.Float64())
			w.Flush() // allows to write buffered output from tabwriter to stdout immediatly
		case newFreq := <-sensorState.SampleFrequencyChangeChan:
			ticker.Stop()
			ticker = time.NewTicker(time.Second / time.Duration(newFreq))
			fmt.Println("Ticker frequency updated to:", newFreq)
		}
	}
}

func publishSensorLog(publishCh *amqp.Channel, sensorname, msg string) error {
	return pubsub.PublishGob(
		publishCh,                // channel
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.BindKeySensorLogs, sensorname), // key
		routing.SensorLog{
			SensorName: sensorname,
			TimeStamp:  time.Now(),
			Message:    msg,
		},
	)
}
