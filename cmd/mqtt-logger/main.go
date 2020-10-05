package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	victron "github.com/christianschmizz/go-victron"
)

const (
	portalID string = "c0847dc9a8cc"
)

func main() {
	brokerIndex := flag.Int("brokerIndex", 0, "Broker's index")
	portalID := flag.String("portalID", "", "portal ID")

	broker := flag.String("broker", "", "The broker URI. ex: tcp://10.10.1.1:1883")

	num := flag.Int("num", 1, "The number of messages to publish or subscribe (default 1)")

	username := flag.String("username", "", "The User (optional)")
	password := flag.String("password", "", "The password (optional)")

	// flag.PrintDefaults()
	flag.Parse()

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	if *broker == "" {
		if *brokerIndex == 0 && *portalID != "" {
			*brokerIndex = victron.BrokerIndexFromPortalID(*portalID)
		}

		if *brokerIndex == 0 {
			log.Fatal().Msgf("invalid broker index: %d", *brokerIndex)
		}

		*broker = victron.BrokerURL(*brokerIndex)
	}

	log.Info().Str("broker", *broker).Msg("using broker")

	conn, err := victron.ConnectBroker(*broker, *username, *password)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect")
	}
	defer conn.Close()

	if err = conn.Subscribe(fmt.Sprintf("N/%s/+/+/#", *portalID)); err != nil {
		log.Fatal().Err(err).Msg("failed to subscribe")
	}

	receiveCount := 0
	for receiveCount < *num {
		incoming := <-conn.Choke
		log.Debug().Str("topic", incoming[0]).Msg(incoming[1])
		receiveCount++
	}

	fmt.Println("Sample Subscriber Disconnected")
}
