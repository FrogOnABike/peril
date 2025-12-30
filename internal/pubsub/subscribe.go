package pubsub

import (
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType string

const (
	Ack         AckType = "ack"
	NackRequeue AckType = "nackrequeue"
	NackDiscard AckType = "nackdiscard"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	// Confirm the channel and queue exist
	ch, qu, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}
	// 	Consume messages
	deliveries, err := ch.Consume(
		qu.Name,
		"",
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}
	// Start a goroutine to handle messages
	go func() {
		for d := range deliveries {
			var val T
			err := json.Unmarshal(d.Body, &val)
			if err != nil {
				continue
			}
			handler(val)
			if err = d.Ack(false); err != nil {
				continue
			}
		}
	}()
	return nil
}
