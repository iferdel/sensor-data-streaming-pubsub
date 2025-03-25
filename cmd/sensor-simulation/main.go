package main

// A client may accidentally or maliciously route messages using non-existent routing keys.
// To avoid complications from lost information, collecting unroutable messages in a RabbitMQ
// alternate exchange is an easy, safe backup. RabbitMQ handles unroutable messages in two ways
// based on the mandatory flag setting within the message header.
// The server either returns the message when the flag is set to "true"
// or silently drops the message when set to "false".
// RabbitMQ let you define an alternate exchange to apply logic to unroutable messages.

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	"github.com/iferdel/sensor-data-streaming-server/internal/sensorlogic"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	rabbitConn *amqp.Connection
	mqttClient mqtt.Client
}

func MQTTCreateClientOptions(clientId, raw string) *mqtt.ClientOptions {
	uri, _ := url.Parse(raw)
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", uri.Host))
	opts.SetUsername(uri.User.Username())
	password, _ := uri.User.Password()
	opts.SetPassword(password)
	opts.SetClientID(clientId)

	return opts
}

func NewConfig() (*Config, error) {
	// amqp
	conn, err := amqp.Dial(routing.RabbitConnString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// mqtt
	mqttOpts := MQTTCreateClientOptions("publisher", routing.RabbitMQTTConnString)
	mqttClient := mqtt.NewClient(mqttOpts)
	token := mqttClient.Connect()
	token.Wait()
	if err := token.Error(); err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ using MQTT connection: %w", err)
	}

	return &Config{
		rabbitConn: conn,
		mqttClient: mqttClient,
	}, nil
}

func main() {

	cfg, err := NewConfig()
	if err != nil {
		log.Fatalf("Could not create rabbitMQ connection: %v", err)
	}
	fmt.Println("Connection to msg broker succeeded")
	defer cfg.rabbitConn.Close()
	defer cfg.mqttClient.Disconnect(200 * uint(time.Millisecond))

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

	cfg.sensorOperation(serialNumber, sampleFrequency)
}

func (cfg *Config) sensorOperation(serialNumber string, sampleFrequency float64) {

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
		pubsub.QueueDurable, // queue duration
		pubsub.QueueClassic, // queue type
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

	// ticker determines how often measurements are read from the original wave (sample rate)
	// so it applies the anti-aliasing (at least two times the signal's maximum frequency)
	ticker := time.NewTicker(time.Second / time.Duration(sensorState.SampleFrequency))
	defer ticker.Stop() // stop Ticker on return so no more ticks will be sent and thus freeing resources

	// batchTimer is the ticker that will trigger the publish of the packet of data
	batchTime := time.Second * 1
	batchTimer := time.NewTicker(batchTime)
	defer batchTimer.Stop()

	sineWaves := []sensorlogic.SineWave{
		{Amplitude: 1.0, Frequency: 5.0, Phase: 0.0},
		{Amplitude: 0.6, Frequency: 2.5, Phase: 0.0},
		{Amplitude: 0.3, Frequency: 60.0, Phase: 0.3},
		{Amplitude: 0.16, Frequency: 120.0, Phase: 0.3},
	}
	startTime := time.Now()
	var measurements []routing.SensorMeasurement

	for {
		select {
		case <-ticker.C:
			accX, timestamp := func() (float64, time.Duration) {
				timestamp := time.Since(startTime)
				elapsedSec := timestamp.Seconds()

				value := sensorlogic.SimulateSignal(sineWaves, elapsedSec)

				return value, timestamp
			}()

			// publish measurement through MQTT
			measurements = append(measurements, routing.SensorMeasurement{
				SerialNumber: serialNumber,
				Timestamp:    startTime.Add(timestamp),
				Value:        accX,
			})

		case <-batchTimer.C:

			if len(measurements) == 0 {
				continue // nothing to send...
			}
			payloadBytes, err := json.Marshal(
				measurements,
			)
			if err != nil {
				log.Printf("Failed to marshal measurements: %v", err)
				return
			}
			pubToken := cfg.mqttClient.Publish(
				fmt.Sprintf(routing.KeySensorMeasurements, serialNumber),
				1,
				true,
				payloadBytes,
			)
			pubToken.Wait()
			if pubToken.Error() != nil {
				log.Printf("Publish error: %v", pubToken.Error())
			}

			measurements = measurements[:0]

		case isSleep := <-sensorState.IsSleepChan:
			if isSleep {
				ticker.Stop()
				batchTimer.Stop()
			} else {
				ticker = time.NewTicker(time.Second / time.Duration(sensorState.SampleFrequency))
				batchTimer = time.NewTicker(batchTime)
			}

		case newFreq := <-sensorState.SampleFrequencyChangeChan:
			ticker.Stop()
			ticker = time.NewTicker(time.Second / time.Duration(newFreq))
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
