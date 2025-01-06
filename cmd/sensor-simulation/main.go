package main

// A client may accidentally or maliciously route messages using non-existent routing keys. To avoid complications from lost information, collecting unroutable messages in a RabbitMQ alternate exchange is an easy, safe backup. RabbitMQ handles unroutable messages in two ways based on the mandatory flag setting within the message header. The server either returns the message when the flag is set to "true" or silently drops the message when set to "false". RabbitMQ let you define an alternate exchange to apply logic to unroutable messages.

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/sensorlogic"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	// environment variables
	serialNumber := os.Getenv("SENSOR_SERIAL_NUMBER")
	if serialNumber == "" {
		fmt.Println("non valid serial number: it is empty")
		return
	}

	sampleFrequencyStr := os.Getenv("SENSOR_SAMPLE_FREQUENCY")
	sampleFrequency, err := strconv.ParseFloat(sampleFrequencyStr, 64)
	if err != nil {
		fmt.Println("non valid sample frequency: it is empty")
		return
	}

	const seed int64 = 99

	sensorOperation(serialNumber, sampleFrequency, seed)
}

func sensorOperation(serialNumber string, sampleFrequency float64, seed int64) {

	bootLogs := []routing.SensorLog{}

	bootLogs = append(bootLogs,
		routing.SensorLog{
			SerialNumber: serialNumber,
			Timestamp:    time.Now(),
			Level:        "INFO",
			Message:      "System powering on...",
		})
	bootLogs = append(bootLogs,
		routing.SensorLog{
			SerialNumber: serialNumber,
			Timestamp:    time.Now(),
			Level:        "INFO",
			Message:      "Bootloader version: v1.0.0",
		})

	conn, err := amqp.Dial(routing.RabbitConnString)
	if err != nil {
		msg := fmt.Sprintf("Could not connect to RabbitMQ: %v", err)
		bootLogs = append(bootLogs,
			routing.SensorLog{
				SerialNumber: serialNumber,
				Timestamp:    time.Now(),
				Level:        "ERROR",
				Message:      msg,
			})
		return
	}
	defer conn.Close()

	bootLogs = append(bootLogs,
		routing.SensorLog{
			SerialNumber: serialNumber,
			Timestamp:    time.Now(),
			Level:        "INFO",
			Message:      "Connection to msg broker succeeded",
		})

	bootLogs = append(bootLogs,
		routing.SensorLog{
			SerialNumber: serialNumber,
			Timestamp:    time.Now(),
			Level:        "INFO",
			Message:      "Loading configuration...",
		})

	sensorState := sensorlogic.NewSensorState(serialNumber, sampleFrequency)

	bootLogs = append(bootLogs,
		routing.SensorLog{
			SerialNumber: serialNumber,
			Timestamp:    time.Now(),
			Level:        "INFO",
			Message:      "Configuration loaded successfully",
		})

	bootLogs = append(bootLogs,
		routing.SensorLog{
			SerialNumber: serialNumber,
			Timestamp:    time.Now(),
			Level:        "INFO",
			Message:      "Performing sensor self-test",
		})
	time.Sleep(500 * time.Millisecond)
	bootLogs = append(bootLogs,
		routing.SensorLog{
			SerialNumber: serialNumber,
			Timestamp:    time.Now(),
			Level:        "INFO",
			Message:      "Self-test result: PASSED",
		})
	time.Sleep(100 * time.Millisecond)

	// publish logic
	publishCh, err := conn.Channel()
	if err != nil {
		msg := fmt.Sprintf("Could not create publish channel: %v", err)
		bootLogs = append(bootLogs,
			routing.SensorLog{
				SerialNumber: serialNumber,
				Timestamp:    time.Now(),
				Level:        "ERROR",
				Message:      msg,
			})
		return
	}

	// Declare an exchange
	err = publishCh.ExchangeDeclare(
		"iot",   // name
		"topic", // type
		true,    // durable
		false,   // auto-deleted
		false,   // internal
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		fmt.Println("Failed to declare an exchange:", err)
		return
	}

	bootLogs = append(bootLogs,
		routing.SensorLog{
			SerialNumber: serialNumber,
			Timestamp:    time.Now(),
			Level:        "INFO",
			Message:      "Publisher channel created",
		})

	// publish sensor for registration if not already
	bootLogs = append(bootLogs,
		routing.SensorLog{
			SerialNumber: serialNumber,
			Timestamp:    time.Now(),
			Level:        "INFO",
			Message:      "Sensor Auth...",
		})
	pubsub.PublishGob(
		publishCh,                // channel
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.KeySensorRegistryFormat, serialNumber)+"."+"created", // routing key
		routing.Sensor{
			SerialNumber: serialNumber,
		}, // based on Data Transfer Object
	)
	// TODO: get back acknowledgment of publish sensor

	// subscribe to sensor command queue
	err = pubsub.SubscribeGob(
		conn,
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.QueueSensorCommandsFormat, serialNumber),       // queue name
		fmt.Sprintf(routing.KeySensorCommandsFormat, serialNumber)+"."+"#", // binding key
		pubsub.QueueDurable, // queue type
		handlerCommand(sensorState),
	)
	if err != nil {
		msg := fmt.Sprintf("Could not subscribe to command: %v", err)
		bootLogs = append(bootLogs,
			routing.SensorLog{
				SerialNumber: serialNumber,
				Timestamp:    time.Now(),
				Level:        "ERROR",
				Message:      msg,
			})
		return
	}

	time.Sleep(100 * time.Millisecond)
	bootLogs = append(bootLogs,
		routing.SensorLog{
			SerialNumber: serialNumber,
			Timestamp:    time.Now(),
			Level:        "INFO",
			Message:      "Successful subscription to iotctl messaging queue",
		})

	bootLogs = append(bootLogs,
		routing.SensorLog{
			SerialNumber: serialNumber,
			Timestamp:    time.Now(),
			Level:        "INFO",
			Message:      "Booting completed, performing measurements...",
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

	show := func(name string, accX any) {
		fmt.Fprintf(w, "%s\t%v\n", name, accX)
	}
	for {
		select {
		case <-ticker.C:
			accX := simulateSignal()
			show(serialNumber, accX)
			w.Flush() // allows to write buffered output from tabwriter to stdout immediatly

			// publish measurement
			pubsub.PublishGob(
				publishCh,
				routing.ExchangeTopicIoT,
				fmt.Sprintf(routing.KeySensorMeasurements, serialNumber),
				routing.SensorMeasurement{
					SerialNumber: serialNumber,
					Timestamp:    time.Now(), // TODO: should it be when the measurement was conceived
					Value:        accX,
				},
			)

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
		fmt.Sprintf(routing.KeySensorLogsFormat, sensorLog.SerialNumber)+"."+"boot", // routing key
		sensorLog, // sensor log
	)
}
