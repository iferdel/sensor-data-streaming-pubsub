package main

// A client may accidentally or maliciously route messages using non-existent routing keys. To avoid complications from lost information, collecting unroutable messages in a RabbitMQ alternate exchange is an easy, safe backup. RabbitMQ handles unroutable messages in two ways based on the mandatory flag setting within the message header. The server either returns the message when the flag is set to "true" or silently drops the message when set to "false". RabbitMQ let you define an alternate exchange to apply logic to unroutable messages.

import (
	"fmt"
	"math"
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
func sensorOperation(wg *sync.WaitGroup, serialNumber string, sampleFrequency float64, seed int64) {
	defer wg.Done() // signals the waitGroup that the goroutine finished its job, bringing the counter down a unit value

	bootLogs := []routing.SensorLog{}

	bootLogs = append(bootLogs,
		routing.SensorLog{
			SensorName: serialNumber,
			TimeStamp:  time.Now(),
			Level:      "INFO",
			Message:    "System powering on...",
		})
	bootLogs = append(bootLogs,
		routing.SensorLog{
			SensorName: serialNumber,
			TimeStamp:  time.Now(),
			Level:      "INFO",
			Message:    "Bootloader version: v1.0.0",
		})

	conn, err := amqp.Dial(routing.RabbitConnString)
	if err != nil {
		msg := fmt.Sprintf("Could not connect to RabbitMQ: %v", err)
		bootLogs = append(bootLogs,
			routing.SensorLog{
				SensorName: serialNumber,
				TimeStamp:  time.Now(),
				Level:      "ERROR",
				Message:    msg,
			})
		return
	}
	defer conn.Close()

	bootLogs = append(bootLogs,
		routing.SensorLog{
			SensorName: serialNumber,
			TimeStamp:  time.Now(),
			Level:      "INFO",
			Message:    "Connection to msg broker succeeded",
		})

	bootLogs = append(bootLogs,
		routing.SensorLog{
			SensorName: serialNumber,
			TimeStamp:  time.Now(),
			Level:      "INFO",
			Message:    "Loading configuration...",
		})

	sensorState := sensorlogic.NewSensorState(serialNumber, sampleFrequency)

	bootLogs = append(bootLogs,
		routing.SensorLog{
			SensorName: serialNumber,
			TimeStamp:  time.Now(),
			Level:      "INFO",
			Message:    "Configuration loaded successfully",
		})

	bootLogs = append(bootLogs,
		routing.SensorLog{
			SensorName: serialNumber,
			TimeStamp:  time.Now(),
			Level:      "INFO",
			Message:    "Performing sensor self-test",
		})
	time.Sleep(500 * time.Millisecond)
	bootLogs = append(bootLogs,
		routing.SensorLog{
			SensorName: serialNumber,
			TimeStamp:  time.Now(),
			Level:      "INFO",
			Message:    "Self-test result: PASSED",
		})
	time.Sleep(100 * time.Millisecond)

	// publish logic
	publishCh, err := conn.Channel()
	if err != nil {
		msg := fmt.Sprintf("Could not create publish channel: %v", err)
		bootLogs = append(bootLogs,
			routing.SensorLog{
				SensorName: serialNumber,
				TimeStamp:  time.Now(),
				Level:      "ERROR",
				Message:    msg,
			})
		return
	}
	bootLogs = append(bootLogs,
		routing.SensorLog{
			SensorName: serialNumber,
			TimeStamp:  time.Now(),
			Level:      "INFO",
			Message:    "Publisher channel created",
		})

	// publish sensor for registration if not already
	bootLogs = append(bootLogs,
		routing.SensorLog{
			SensorName: serialNumber,
			TimeStamp:  time.Now(),
			Level:      "INFO",
			Message:    "Sensor Auth...",
		})
	pubsub.PublishGob(
		publishCh,                // channel
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.KeySensorRegistryFormat, serialNumber)+"."+"created", // routing key
		routing.Sensor{
			SensorName: serialNumber,
		}, // based on Data Transfer Object
	)
	// get back acknowledgment of publish sensor

	// subscribe to sensor command queue
	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.QueueSensorCommandsFormat, serialNumber), // queue name
		// fmt.Sprintf(routing.BindKeySensorCommandFormat, serialNumber), // binding key
		fmt.Sprintf(routing.KeySensorCommandsFormat, serialNumber)+"."+"#", // binding key
		pubsub.QueueDurable, // queue type
		handlerCommand(sensorState),
	)
	if err != nil {
		msg := fmt.Sprintf("Could not subscribe to command: %v", err)
		bootLogs = append(bootLogs,
			routing.SensorLog{
				SensorName: serialNumber,
				TimeStamp:  time.Now(),
				Level:      "ERROR",
				Message:    msg,
			})
		return
	}

	time.Sleep(100 * time.Millisecond)
	bootLogs = append(bootLogs,
		routing.SensorLog{
			SensorName: serialNumber,
			TimeStamp:  time.Now(),
			Level:      "INFO",
			Message:    "Successful subscription to iotctl messaging queue",
		})

	bootLogs = append(bootLogs,
		routing.SensorLog{
			SensorName: serialNumber,
			TimeStamp:  time.Now(),
			Level:      "INFO",
			Message:    "Booting completed, performing measurements...",
		})

	// publish sensor boot logs
	for _, bootLog := range bootLogs {
		err = publishSensorLog(
			publishCh,
			bootLog,
		)
		if err != nil {
			fmt.Printf("Error publishing log: %s\n", err)
		}
	}

	ticker := time.NewTicker(time.Second / time.Duration(sensorState.SampleFrequency))
	defer ticker.Stop() // stop Ticker on return so no more ticks will be sent and thus freeing resources

	var t float64 // will track time to ensure the running phase of the sine wave
	_ = rand.New(rand.NewSource(seed))
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)

	simulateSignal := func() float64 {
		// vibration parameters
		amplitude := 1.0
		freq := sensorState.SampleFrequency
		dt := 1.0 / freq // time between measurements

		// increment time with each call (tracking running phase)
		t += dt

		// angular frequency (constant based on the sinewave frequency)
		w := 2 * math.Pi * freq

		// sine wave
		return amplitude * math.Sin(w*t)
	}

	show := func(name string, accX, accY, accZ any) {
		fmt.Fprintf(w, "%s\t%v\t%v\t%v\n", name, accX, accY, accZ)
	}
	for {
		select {
		case <-ticker.C:
			accX := simulateSignal()
			accY := simulateSignal()
			accZ := simulateSignal()
			show(serialNumber, accX, accY, accZ)
			w.Flush() // allows to write buffered output from tabwriter to stdout immediatly
		case newFreq := <-sensorState.SampleFrequencyChangeChan:
			ticker.Stop()
			ticker = time.NewTicker(time.Second / time.Duration(newFreq))
			fmt.Println("Ticker frequency updated to:", newFreq)
		}
	}
}

func publishSensorLog(publishCh *amqp.Channel, sensorLog routing.SensorLog) error {
	return pubsub.PublishGob(
		publishCh,                // channel
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.KeySensorLogsFormat, sensorLog.SensorName)+"."+"boot", // routing key
		sensorLog, // sensor log
	)
}
