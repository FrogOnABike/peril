package main

import (
	"fmt"
	"time"

	"github.com/frogonabike/peril/internal/gamelogic"
	"github.com/frogonabike/peril/internal/pubsub"
	"github.com/frogonabike/peril/internal/routing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
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
			if err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, mv.Player.Username),
				gamelogic.RecognitionOfWar{
					Attacker: mv.Player,
					Defender: gs.GetPlayerSnap(),
				},
			); err != nil {
				fmt.Println("Failed to publish war recognition:", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			return pubsub.NackDiscard
		}

	}
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(war gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")
		wo, wp, lp := gs.HandleWar(war)
		switch wo {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.Ack
		case gamelogic.WarOutcomeYouWon:
			fmt.Println("Congratulations! You won the war!")
			glMsg := routing.GameLog{
				Username:    gs.Player.Username,
				Message:     fmt.Sprintf("%s has won a war against %s", wp, lp),
				CurrentTime: time.Now(),
			}
			err := pubsub.PublishGameLog(glMsg, ch)
			if err != nil {
				fmt.Println("Failed to publish game log:", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeOpponentWon:
			fmt.Println("You lost the war. Better luck next time.")
			glMsg := routing.GameLog{
				Username:    gs.Player.Username,
				Message:     fmt.Sprintf("%s has won a war against %s", wp, lp),
				CurrentTime: time.Now(),
			}
			err := pubsub.PublishGameLog(glMsg, ch)
			if err != nil {
				fmt.Println("Failed to publish game log:", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		case gamelogic.WarOutcomeDraw:
			fmt.Println("The war ended in a draw.")
			glMsg := routing.GameLog{
				Username:    gs.Player.Username,
				Message:     fmt.Sprintf("A war between %s and %s ended in a draw.", wp, lp),
				CurrentTime: time.Now(),
			}
			err := pubsub.PublishGameLog(glMsg, ch)
			if err != nil {
				fmt.Println("Failed to publish game log:", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		default:
			return pubsub.NackDiscard
		}
	}
}
