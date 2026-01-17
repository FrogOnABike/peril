package main

import (
	"fmt"

	"github.com/frogonabike/peril/internal/gamelogic"
	"github.com/frogonabike/peril/internal/pubsub"
	"github.com/frogonabike/peril/internal/routing"
)

func handleGameLog(log routing.GameLog) pubsub.AckType {
	defer fmt.Print("> ")
	err := gamelogic.WriteLog(log)
	if err != nil {
		fmt.Println("Failed to write game log:", err)
		return pubsub.NackRequeue
	}
	fmt.Printf("Game log recorded: %s\n", log.Message)
	return pubsub.Ack

}
