package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"

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
	return subscribe(conn, exchange, queueName, key, queueType, handler, func(body []byte) (T, error) {
		var val T
		err := json.Unmarshal(body, &val)
		return val, err
	})
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	return subscribe(conn, exchange, queueName, key, queueType, handler, func(body []byte) (T, error) {
		var val T
		dec := gob.NewDecoder(bytes.NewReader(body))
		err := dec.Decode(&val)
		return val, err
	})
}

// // Confirm the channel and queue exist
// ch, qu, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
// if err != nil {
// 	return err
// }
// // 	Consume messages
// deliveries, err := ch.Consume(
// 	qu.Name,
// 	"",
// 	false, // auto-ack
// 	false, // exclusive
// 	false, // no-local
// 	false, // no-wait
// 	nil,   // args
// )
// if err != nil {
// 	return err
// }
// // Start a goroutine to handle messages
// go func() {
// 	for d := range deliveries {
// 		var val T
// 		err := json.Unmarshal(d.Body, &val)
// 		if err != nil {
// 			continue
// 		}
// 		at := handler(val)
// 		switch at {
// 		case Ack:
// 			if err = d.Ack(false); err != nil {
// 				fmt.Printf("could not ack message: %v\n", err)
// 				continue
// 			}
// 			fmt.Printf("acked message\n")
// 		case NackRequeue:
// 			if err = d.Nack(false, true); err != nil {
// 				fmt.Printf("could not nackrequeue message: %v\n", err)
// 				continue
// 			}
// 			fmt.Printf("nackrequeued message\n")
// 		case NackDiscard:
// 			if err = d.Nack(false, false); err != nil {
// 				fmt.Printf("could not nackdiscard message: %v\n", err)
// 				continue
// 			}
// 			fmt.Printf("nackdiscarded message\n")
// 		default:
// 			fmt.Printf("unknown AckType: %v\n", at)
// 		}
// 	}
// }()
// return nil
// }

func subscribe[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType SimpleQueueType,
	handler func(T) AckType,
	unmarshaller func([]byte) (T, error),
) error {

	// Confirm the channel and queue exist
	ch, qu, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
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
			val, err := unmarshaller(d.Body)
			if err != nil {
				continue
			}
			at := handler(val)
			switch at {
			case Ack:
				if err = d.Ack(false); err != nil {
					fmt.Printf("could not ack message: %v\n", err)
					continue
				}
				fmt.Printf("acked message\n")
			case NackRequeue:
				if err = d.Nack(false, true); err != nil {
					fmt.Printf("could not nackrequeue message: %v\n", err)
					continue
				}
				fmt.Printf("nackrequeued message\n")
			case NackDiscard:
				if err = d.Nack(false, false); err != nil {
					fmt.Printf("could not nackdiscard message: %v\n", err)
					continue
				}
				fmt.Printf("nackdiscarded message\n")
			default:
				fmt.Printf("unknown AckType: %v\n", at)
			}
		}
	}()
	return nil
}
