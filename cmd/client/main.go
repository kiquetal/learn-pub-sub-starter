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
	fmt.Println("Starting Peril client...")
	fmt.Println("Connecting to RabbitMQ...")
	connUrl := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connUrl)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ")
		panic(err)
	}
	name, err := gamelogic.ClientWelcome()

	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Println("Failed to close connection")
		}
	}()

	_, queue, errorBinding := pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, fmt.Sprintf("%s.%s", routing.PauseKey, name), routing.PauseKey, pubsub.TransientQueue)
	if errorBinding != nil {
		fmt.Println("Failed to declare and bind queue")
		panic(err)
	}

	gs := gamelogic.NewGameState(name)

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			if len(words) != 3 {
				fmt.Println("Invalid spawn command")
				continue
			}
			err := gs.CommandSpawn(words)
			if err != nil {
				fmt.Println(err)
			}
		case "move":
			if len(words) < 3 {
				fmt.Println("Invalid move command")
				continue

			}
			_, err := gs.CommandMove(words)
			if err != nil {
				fmt.Println(err)

			}
			fmt.Println("Move command executed")
		}

	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)

	sigs := <-sig
	fmt.Println("Received signal: ", sigs)

	fmt.Println("Queue created: ", queue.Name)

}
