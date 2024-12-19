package cmd

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "iotctl",
	Short: "CLI Tool for Interacting with IoT sensors using pubsub system",
	Long: `This CLI tool allows you to manage resources (sensors) 
It allows the use of keywords to alter the behaviour of the available sensors in the cluster. 
Every command has some flags such as sensor id or parameters related to the command itself.`,
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
}

var (
	conn      *amqp.Connection
	publishCh *amqp.Channel
	err       error
)

func initConfig() {

	const rabbitConnString = "amqp://guest:guest@localhost:5672/"
	conn, err = amqp.Dial(rabbitConnString)
	if err != nil {
		log.Fatalf("could not connect to RabbitMQ: %v", err)
	}

	fmt.Println("Server connected to RabbitMQ")
	publishCh, err = conn.Channel()
	if err != nil {
		log.Fatalf("could not create channel: %v", err)
	}
}
