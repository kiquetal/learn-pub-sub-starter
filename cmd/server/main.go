package main

import (
	"fmt"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"os/signal"
	"syscall"
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

	defer conn.Close()

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
			break
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
