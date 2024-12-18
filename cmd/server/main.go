package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/iferdel/sensor-data-streaming-server/internal/pubsub"
	"github.com/iferdel/sensor-data-streaming-server/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	const rabbitConnString = "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}
	defer conn.Close()
	fmt.Println("Server connected to RabbitMQ")

	publishCh, err := conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}

	sensorSerialNumber := "AAD-1123"

	for {
		commandInput := GetInput()
		if len(commandInput) == 0 {
			continue
		}
		switch commandInput[0] {
		case "sleep":
			fmt.Println("sending sleep command to sensor", sensorSerialNumber)
			err = pubsub.PublishGob(
				publishCh,                // amqp.Channel
				routing.ExchangeTopicIoT, // exchange
				fmt.Sprintf(routing.BindKeySensorCommandFormat, sensorSerialNumber), // routing key
				routing.CommandMessage{
					SensorName: sensorSerialNumber,
					Timestamp:  time.Now(),
					Command:    "sleep",
					Params:     nil,
				}, // value
			)
			if err != nil {
				log.Printf("could not publish sleep command: %v", err)
			}
		case "resume":
			fmt.Println("sending resume command to sensor", sensorSerialNumber)
			err = pubsub.PublishGob(
				publishCh,                // amqp.Channel
				routing.ExchangeTopicIoT, // exchange
				fmt.Sprintf(routing.BindKeySensorCommandFormat, sensorSerialNumber), // routing key
				routing.CommandMessage{
					SensorName: sensorSerialNumber,
					Timestamp:  time.Now(),
					Command:    "resume",
					Params:     nil,
				}, // value
			)
			if err != nil {
				log.Printf("could not publish resume command: %v", err)
			}
		case "changeSampleFrequency":
			if commandInput[1] == "" {
				fmt.Println("sample frequency must have an argument for the value")
				continue
			}
			value, err := strconv.Atoi(commandInput[1])
			if err != nil {
				fmt.Println("sample frequency must be an integer greater than 0")
				continue
			}
			fmt.Println("sending change sample frequency command to sensor", sensorSerialNumber)
			err = pubsub.PublishGob(
				publishCh,                // amqp.Channel
				routing.ExchangeTopicIoT, // exchange
				fmt.Sprintf(routing.BindKeySensorCommandFormat, sensorSerialNumber), // routing key
				routing.CommandMessage{
					SensorName: sensorSerialNumber,
					Timestamp:  time.Now(),
					Command:    "changeSampleFrequency",
					Params: map[string]interface{}{
						"sampleFrequency": value,
					},
				}, // value
			)
			if err != nil {
				log.Printf("could not publish change sample frequency command: %v", err)
			}
		}
	}
}

func GetInput() []string {
	fmt.Print("> ")
	scanner := bufio.NewScanner(os.Stdin)
	scanned := scanner.Scan()
	if !scanned {
		return nil
	}
	line := scanner.Text()
	line = strings.TrimSpace(line)
	return strings.Fields(line)
}
