package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

const DurableQueue = 0
const TransientQueue = 1

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
func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int, callback func(T)) error {

	channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		fmt.Println("Failed to declare and bind queue")
		return err

	}
	subscriptos, err := channel.Consume(queue.Name, "", false, false, false, false, nil)

	if err != nil {
		fmt.Println("Failed to consume the queue")
		return err
	}

	go func() {
		for msg := range subscriptos {
			var data T
			err := json.Unmarshal(msg.Body, &data)
			if err != nil {
				fmt.Println("Failed to unmarshal the message")
			}
			callback(data)
			msg.Ack(false)
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
