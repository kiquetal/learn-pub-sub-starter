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

var positionMap map[string]string = map[string]string{
	"americas":   "americas",
	"europe":     "europe",
	"asia":       "asia",
	"africa":     "africa",
	"antarctica": "antarctica",
	"australia":  "australia",
}

var unitTypeMap map[string]string = map[string]string{
	"infantry":  "infantry",
	"cavalry":   "cavalry",
	"artillery": "artillery",
}

func main() {
	fmt.Println("Starting Peril client...")
	fmt.Println("Connecting to RabbitMQ...")
	connUrl := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connUrl)
	if err != nil {
		fmt.Println("Failed to connect to RabbitMQ")
		panic(err)
	}
	channel, err := conn.Channel()
	if err != nil {
		fmt.Println("Failed to open a channel")
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

	pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, fmt.Sprintf(routing.GameLogSlug), fmt.Sprintf("game_logs.*"), pubsub.DurableQueue)
	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, fmt.Sprintf("pause.%s", name), routing.PauseKey, pubsub.TransientQueue, handlerPause(gs))
	if err != nil {
		fmt.Println("Failed to subscribe to pause")
		panic(err)

	}
	pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.ArmyMovesPrefix, name), fmt.Sprintf("%s.*", routing.ArmyMovesPrefix), pubsub.TransientQueue, handlerMove(gs))

myloop:
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			err := gs.CommandSpawn(words)
			if err != nil {
				fmt.Println(err)
			}
		case "move":
			movement, err := gs.CommandMove(words)
			if err != nil {
				fmt.Println(err)
			}
			pubsub.PublishJSON(channel, routing.ExchangePerilTopic, fmt.Sprintf("army_moves.%s", name), movement)
		case "status":
			gs.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			break myloop
		default:
			fmt.Println("Unknown command")
		}

	}
	fmt.Println("Output game")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT)

	sigs := <-sig
	fmt.Println("Received signal: ", sigs)

	fmt.Println("Queue created: ", queue.Name)

}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Println("> ")
		gs.HandlePause(ps)
		return pubsub.Ack

	}
}
func handlerMove(gs *gamelogic.GameState) func(state gamelogic.ArmyMove) pubsub.AckType {
	return func(mc gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Println("> ")
		rt := gs.HandleMove(mc)
		switch rt {
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			return pubsub.Ack
		default:
			return pubsub.NackDiscard

		}

	}
}
