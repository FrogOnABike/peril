package main

import (
	"fmt"
	"os"
	"os/signal"

	// "github.com/frogonabike/peril/internal/gamelogic"
	// "github.com/frogonabike/peril/internal/pubsub"

	"github.com/frogonabike/peril/internal/gamelogic"
	"github.com/frogonabike/peril/internal/pubsub"
	"github.com/frogonabike/peril/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")

	connectString := "amqp://guest:guest@localhost:5672/"
	fmt.Println("Starting Peril client...")
	conn, err := amqp.Dial(connectString)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ:", err)
		return
	}
	defer conn.Close()
	fmt.Println("Connection successful")

	// Prompt for username
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Declare and bind queues
	chan1, _, err := pubsub.DeclareAndBind(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.Transient,
	)
	if err != nil {
		fmt.Println("Failed to declare and bind queue:", err)
		return
	}
	defer chan1.Close()

	// Create a new game state
	gameState := gamelogic.NewGameState(username)

	// Subscribe to pause messages
	if err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilDirect,
		fmt.Sprintf("%s.%s", routing.PauseKey, username),
		routing.PauseKey,
		pubsub.Transient,
		handlerPause(gameState)); err != nil {
		fmt.Println("Failed to subscribe to pause messages:", err)
		return
	}

	// Subscribe to army move messages
	if err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		fmt.Sprintf("army_moves.%s", username),
		"army_moves.*",
		pubsub.Transient,
		handlerMove(gameState, chan1)); err != nil {
		fmt.Println("Failed to subscribe to army move messages:", err)
		return
	}

	// Subscribe to war messages
	if err = pubsub.SubscribeJSON(
		conn,
		routing.ExchangePerilTopic,
		"war",
		fmt.Sprintf("%s.*", routing.WarRecognitionsPrefix),
		pubsub.Durable,
		handlerWar(gameState, chan1)); err != nil {
		fmt.Println("Failed to subscribe to war messages:", err)
		return
	}

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
			case "spawn":
				if err := gameState.CommandSpawn(words); err != nil {
					fmt.Println(err)
				}
			case "move":
				mv, err := gameState.CommandMove(words)
				if err != nil {
					fmt.Println(err)
				}
				if err = pubsub.PublishJSON(
					chan1,
					routing.ExchangePerilTopic,
					fmt.Sprintf("army_moves.%s", username),
					mv); err != nil {
					fmt.Println("Failed to publish army move message:", err)
				} else {
					fmt.Println("Army move message published")
				}

			case "status":
				gameState.CommandStatus()
			case "spam":
				fmt.Println("Spamming not allowed yet")
			case "help":
				gamelogic.PrintClientHelp()
			case "quit":
				gamelogic.PrintQuit()
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
	fmt.Println("Shutting down Peril client...")
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
}
