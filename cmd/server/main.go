package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/frogonabike/peril/internal/gamelogic"
	"github.com/frogonabike/peril/internal/pubsub"
	"github.com/frogonabike/peril/internal/routing"
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

	// Open a channel
	chan1, err := conn.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel:", err)
		return
	}
	defer chan1.Close()
	fmt.Println("Channel opened successfully")

	// Display help message
	gamelogic.PrintServerHelp()

	// prepare channels for input and signals
	inputCh := make(chan []string)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	// reader goroutine: push parsed words to inputCh
	go func() {
		for {
			words := gamelogic.GetInput()
			inputCh <- words
		}
	}()

	// main loop: handle input or OS signal
	for {
		select {
		case words := <-inputCh:
			if len(words) == 0 {
				continue
			}
			cmd := words[0]
			switch cmd {
			case "pause":
				if err := pubsub.PublishJSON(chan1, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true}); err != nil {
					fmt.Println("Failed to publish a pause message:", err)
				} else {
					fmt.Println("Pause message published")
				}
			case "resume":
				if err := pubsub.PublishJSON(chan1, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false}); err != nil {
					fmt.Println("Failed to publish a resume message:", err)
				} else {
					fmt.Println("Resume message published")
				}
			case "quit":
				fmt.Println("Quitting server...")
				// perform any cleanup if needed, then exit loop to reach shutdown logic
				goto shutdown
			default:
				fmt.Println("Unknown command. Type 'help' for a list of commands.")
			}
		case <-signalChan:
			fmt.Println("Received interrupt signal")
			goto shutdown
		}
	}

shutdown:
	fmt.Println("Shutting down Peril server...")
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
}
