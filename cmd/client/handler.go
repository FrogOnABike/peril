package main

import (
	"fmt"

	"github.com/frogonabike/peril/internal/gamelogic"
	"github.com/frogonabike/peril/internal/pubsub"
	"github.com/frogonabike/peril/internal/routing"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(mv gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		mo := gs.HandleMove(mv)
		switch mo {
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
			// do nothing
		case gamelogic.MoveOutComeSafe:
			fmt.Println("Your units are safe.")
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			fmt.Println("Prepare for battle!")
			return pubsub.Ack
		default:
			return pubsub.NackDiscard
		}

	}
}
