package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

	conn_url := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(conn_url)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ")
		panic(err)
	}
	fmt.Println("Starting Peril server...")

	channel, err := conn.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel")
		panic(err)
	}

	fmt.Println("Publishing pause message...")

	pubsub.SubscribeGeneric(conn, routing.GameLogSlug, routing.GameLogSlug, "games_logs.*", pubsub.DurableQueue, handlerLogs, decodeGob)

	defer conn.Close()
mainLoop:
	for {
		gamelogic.PrintServerHelp()
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "pause":
			pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: true,
			})
		case "resume":
			pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{
				IsPaused: false,
			})
		case "quit":
			break mainLoop
		default:
			fmt.Println("Unknown command")

		}

	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)

	sigs := <-sig
	fmt.Println("Received signal: ", sigs)
	fmt.Println("Closing Peril server...")

}
func handlerLogs(data string) pubsub.AckType {

	defer fmt.Println("> ")
	var log routing.GameLog
	log = routing.GameLog{
		CurrentTime: time.Now(),
		Message:     data,
		Username:    "server",
	}

	gamelogic.WriteLog(log)

	return pubsub.Ack
}

func decodeGob(data []byte) (string, error) {
	var message string
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&message)
	if err != nil {
		return "", err
	}
	return message, nil

}
