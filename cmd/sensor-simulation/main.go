package main

// A client may accidentally or maliciously route messages using non-existent routing keys. To avoid complications from lost information, collecting unroutable messages in a RabbitMQ alternate exchange is an easy, safe backup. RabbitMQ handles unroutable messages in two ways based on the mandatory flag setting within the message header. The server either returns the message when the flag is set to "true" or silently drops the message when set to "false". RabbitMQ let you define an alternate exchange to apply logic to unroutable messages.

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/sensorlogic"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	rabbitConn *amqp.Connection
}

func NewConfig() (*Config, error) {
	conn, err := amqp.Dial(routing.RabbitConnString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return &Config{
		rabbitConn: conn,
	}, nil
}

func main() {

	cfg, err := NewConfig()
	if err != nil {
		log.Fatalf("Could not create rabbitMQ connection: %v", err)
	}
	fmt.Println("Connection to msg broker succeeded")
	defer cfg.rabbitConn.Close()

	// environment variables
	serialNumber := os.Getenv("SENSOR_SERIAL_NUMBER")
	if serialNumber == "" {
		log.Fatal("non valid serial number: it is empty")
	}

	sampleFrequencyStr := os.Getenv("SENSOR_SAMPLE_FREQUENCY")
	sampleFrequency, err := strconv.ParseFloat(sampleFrequencyStr, 64)
	if err != nil {
		log.Fatal("non valid sample frequency: it is empty")
	}

	const seed int64 = 99

	cfg.sensorOperation(serialNumber, sampleFrequency, seed)
}

func (cfg *Config) sensorOperation(serialNumber string, sampleFrequency float64, seed int64) {

	sensorState := sensorlogic.NewSensorState(serialNumber, sampleFrequency)
	// publish logic
	publishCh, err := cfg.rabbitConn.Channel()
	// this should be written to /var/log as a log that is in the sensor (and not published)
	if err != nil {
		log.Fatalf("Could not create publish channel: %v", err)
	}
	fmt.Println("Publisher channel created")

	// goroutine for sensor logs publish
	go func() {
		for {
			select {
			case infoMsg := <-sensorState.LogsInfo:
				publishSensorLog(publishCh, routing.SensorLog{
					SerialNumber: serialNumber,
					Timestamp:    time.Now(),
					Level:        "INFO",
					Message:      infoMsg,
				})
			case warningMsg := <-sensorState.LogsWarning:
				publishSensorLog(publishCh, routing.SensorLog{
					SerialNumber: serialNumber,
					Timestamp:    time.Now(),
					Level:        "WARNING",
					Message:      warningMsg,
				})
			case errMsg := <-sensorState.LogsError:
				publishSensorLog(publishCh, routing.SensorLog{
					SerialNumber: serialNumber,
					Timestamp:    time.Now(),
					Level:        "ERROR",
					Message:      errMsg,
				})
			}
		}
	}()

	sensorState.LogsInfo <- "System powering on..."
	time.Sleep(100 * time.Millisecond)
	sensorState.LogsInfo <- "Bootloader version: v1.0.0"
	time.Sleep(200 * time.Millisecond)
	sensorState.LogsInfo <- "Loading configuration..."
	sensorState.LogsInfo <- "Configuration loaded successfully"
	time.Sleep(100 * time.Millisecond)
	sensorState.LogsInfo <- "Performing sensor self-test"
	time.Sleep(500 * time.Millisecond)
	sensorState.LogsInfo <- "Self-test result: PASSED"
	time.Sleep(100 * time.Millisecond)

	// publish sensor for registration if not already
	sensorState.LogsInfo <- "Sensor Auth..."
	pubsub.PublishGob(
		publishCh,                // channel
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.KeySensorRegistryFormat, serialNumber)+"."+"created", // routing key
		routing.Sensor{
			SerialNumber:    serialNumber,
			SampleFrequency: sampleFrequency,
		}, // based on Data Transfer Object
	)
	// TODO: get back acknowledgment of publish sensor

	// subscribe to sensor command queue
	err = pubsub.SubscribeGob(
		cfg.rabbitConn,
		routing.ExchangeTopicIoT, // exchange
		fmt.Sprintf(routing.QueueSensorCommandsFormat, serialNumber),       // queue name
		fmt.Sprintf(routing.KeySensorCommandsFormat, serialNumber)+"."+"#", // binding key
		pubsub.QueueDurable, // queue type
		handlerCommand(sensorState),
	)
	if err != nil {
		sensorState.LogsError <- fmt.Sprintf("Could not subscribe to command queue: %v\n", err)
		return
	}
	time.Sleep(100 * time.Millisecond)
	sensorState.LogsInfo <- "Successful subscription to iotctl messaging queue"

	time.Sleep(500 * time.Millisecond)
	sensorState.LogsInfo <- "Booting completed, performing measurements..."

	ticker := time.NewTicker(time.Second / time.Duration(sensorState.SampleFrequency))
	defer ticker.Stop() // stop Ticker on return so no more ticks will be sent and thus freeing resources

	_ = rand.New(rand.NewSource(seed))
	w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)

	// Example parameters
	amplitude1 := 1.0
	frequency1 := 5.0 // 5 Hz
	amplitude2 := 0.6
	frequency2 := 2.5 // 2.5 Hz

	// Initialize sine waves
	sine1 := sensorlogic.SineWave{
		Amplitude: amplitude1,
		Frequency: frequency1,
		Phase:     0.0,
	}
	sine2 := sensorlogic.SineWave{
		Amplitude: amplitude2,
		Frequency: frequency2,
		Phase:     0.0,
	}
	var dt float64 = 0.0 // Tracks the elapsed time in seconds
	dtIncrement := 1.0 / sensorState.SampleFrequency

	// Mutex to protect dt in case of concurrent access
	var mu sync.Mutex

	// Function to simulate a single sample
	simulateSample := func() float64 {
		mu.Lock()
		defer mu.Unlock()
		// Generate the superimposed signal
		value := sine1.Generate(dt) + sine2.Generate(dt)
		// Increment time
		dt += dtIncrement
		return value
	}

	show := func(name string, accX any) {
		fmt.Fprintf(w, "%s\t%v\n", name, accX)
	}
	for {
		select {
		case <-ticker.C:
			accX := simulateSample()
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
		fmt.Sprintf(routing.KeySensorLogsFormat, sensorLog.SerialNumber), // routing key
		sensorLog, // sensor log
	)
}
