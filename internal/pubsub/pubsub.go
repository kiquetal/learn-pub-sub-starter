package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

const DurableQueue = 0
const TransientQueue = 1

type AckType string

const (
	Ack         AckType = "Ack"
	NackRequeue AckType = "NackRequeue"
	NackDiscard AckType = "NackDiscard"
)

func PublishGob[T any](ch *amqp.Channel, exchange, key string, data T) error {

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	if err != nil {
		return err

	}

	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        buf.Bytes(),
	})

}

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, data T) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        jsonData,
	})
}
func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int, callback func(T) AckType) error {

	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		fmt.Println("Failed to declare and bind queue")
		return err

	}

	subscriptions, err := channel.Consume(queue.Name, "", false, false, false, false, nil)

	if err != nil {
		fmt.Println("Failed to consume the queue")
		return err
	}

	go func() {
		for msg := range subscriptions {
			var data T
			err := json.Unmarshal(msg.Body, &data)
			if err != nil {
				fmt.Println("Failed to unmarshal the message")
			}
			result := callback(data)

			switch result {
			case Ack:
				fmt.Println("[Ack]")
				msg.Ack(false)
			case NackRequeue:
				fmt.Println("[NackRequeue]")
				msg.Nack(false, true)
			case NackDiscard:
				fmt.Println("[NackDiscard]")
				msg.Nack(false, false)

			default:
				fmt.Println("[NackRequeue]-default")

				msg.Nack(false, true)
			}

		}
	}()

	return nil
}

func DeclareAndBindNotDLQ(conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int) (*amqp.Channel, amqp.Queue, error) {

	channel, err := conn.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel")
		return nil, amqp.Queue{}, err
	}

	var queue amqp.Queue

	switch simpleQueueType {
	case DurableQueue:
		queue, err = channel.QueueDeclare(queueName, true, false, false, false, nil)
	case TransientQueue:
		queue, err = channel.QueueDeclare(queueName, false, true, true, false, nil)
	}

	if err != nil {
		fmt.Printf("Failed to declare queue: %v\\n", err)
		return nil, amqp.Queue{}, err
	}

	err = channel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		fmt.Printf("Failed to bind queue: %v\\n", err)
		return nil, amqp.Queue{}, err
	}

	return channel, queue, nil

}

func SubscribeGeneric[T any](conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int, callback func(T) AckType, unmarshaller func([]byte) (T, error)) error {

	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		fmt.Println("Failed to declare and bind queue")
		return err

	}
	channel.Qos(10, 0, false)
	subscriptions, err := channel.Consume(queue.Name, "", false, false, false, false, nil)

	if err != nil {
		fmt.Println("Failed to consume the queue")
		return err
	}

	go func() {
		for msg := range subscriptions {
			var data T
			data, err := unmarshaller(msg.Body)

			if err != nil {
				fmt.Println("Failed to unmarshal the message")
			}

			if err != nil {
				fmt.Println("Failed to unmarshal the message")
			}
			result := callback(data)

			switch result {
			case Ack:
				fmt.Println("[Ack]")
				msg.Ack(false)
			case NackRequeue:
				fmt.Println("[NackRequeue]")
				msg.Nack(false, true)
			case NackDiscard:
				fmt.Println("[NackDiscard]")
				msg.Nack(false, false)

			default:
				fmt.Println("[NackRequeue]-default")

				msg.Nack(false, true)
			}

		}
	}()

	return nil
}

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int) (*amqp.Channel, amqp.Queue, error) {

	channel, err := conn.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel")
		return nil, amqp.Queue{}, err
	}

	var queue amqp.Queue

	var table amqp.Table
	table = amqp.Table{
		"x-dead-letter-exchange": "peril_dlx",
	}

	switch simpleQueueType {
	case DurableQueue:
		queue, err = channel.QueueDeclare(queueName, true, false, false, false, table)
	case TransientQueue:
		queue, err = channel.QueueDeclare(queueName, false, true, true, false, table)
	}

	if err != nil {
		fmt.Printf("Failed to declare queue: %v\\n", err)
		return nil, amqp.Queue{}, err
	}

	err = channel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		fmt.Printf("Failed to bind queue: %v\\n", err)
		return nil, amqp.Queue{}, err
	}

	return channel, queue, nil

}
