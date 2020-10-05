package vrm

import (
	"crypto/tls"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
)

const (
	clientID string = "govictrongo"
)

func BrokerIndexFromPortalID(portalID string) int {
	sum := 0
	for _, char := range portalID {
		sum += int(char)
	}
	return sum % 128
}

func BrokerURL(brokerIndex int) string {
	return fmt.Sprintf("tcps://mqtt%d.victronenergy.com:8883", brokerIndex)
}

type brokerConnection struct {
	client mqtt.Client
	Choke  chan [2]string
	topics []string
}

func ConnectBroker(broker, username, password string) (*brokerConnection, error) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetCleanSession(false)
	opts.SetKeepAlive(2 * time.Second)
	opts.SetPingTimeout(1 * time.Second)
	opts.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})

	conn := brokerConnection{
		Choke: make(chan [2]string),
	}

	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		conn.Choke <- [2]string{msg.Topic(), string(msg.Payload())}
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("failed to connect: %w", token.Error())
	}

	conn.client = client

	return &conn, nil
}

func (c *brokerConnection) Subscribe(topic string) error {
	if token := c.client.Subscribe(topic, byte(0), nil); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe: %w", token.Error())
	}
	c.topics = append(c.topics, topic)
	return nil
}

func (c *brokerConnection) Close() {
	for _, topic := range c.topics {
		if token := c.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			log.Error().Err(token.Error()).Str("topic", topic).Msg("failed to unsubscribe when closing")
		}
	}
	c.client.Disconnect(250)
}
