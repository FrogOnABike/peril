package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/FrogOnABike/peril/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	connectString := "amqp://guest:guest@localhost:5672/"
	fmt.Println("Starting Peril server...")
	conn, err := amqp.Dial(connectString)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ:", err)
		return
	}
	defer conn.Close()
	fmt.Println("Connection successful")

	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("Shutting down Peril server...")

	// Open a channel
	chan1, err := conn.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel:", err)
		return
	}
	defer chan1.Close()
	fmt.Println("Channel opened successfully")

	// Publish a message to the exchange
	pubsub.PublishJSON(chan1, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
}
