package pubsub

import (
	"context"
	"encoding/json"

	"github.com/frogonabike/peril/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

type SimpleQueueType string

const (
	Durable   SimpleQueueType = "durable"
	Transient SimpleQueueType = "transient"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	valBytes, err := json.Marshal(val)
	if err != nil {
		return err
	}
	ch.PublishWithContext(context.Background(),
		exchange,
		key,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        valBytes,
		},
	)
	return nil
}

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	// Open a channel
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	// Declare the queue
	qu, err := ch.QueueDeclare(
		queueName,
		queueType == Durable,   // durable
		queueType == Transient, // autodelete
		queueType == Transient, // exclusive
		false,                  // no-wait
		amqp.Table{
			"x-dead-letter-exchange": routing.ExchangePerilDLX,
		},
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	// Bind the queue to the exchange with the routing key
	err = ch.QueueBind(
		qu.Name,
		key,
		exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	return ch, qu, nil
}
