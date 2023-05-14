package mqtt

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/wamphlett/nv7-pi-controller/config"
	"github.com/wamphlett/nv7-pi-controller/pkg/controller"
)

// payload represents the JSON payload which is published
type payload struct {
	Button  string
	Channel string
	Theme   string
	Colour  string
	Speed   int
}

// Publisher defines the publisher methods
type Publisher struct {
	client mqtt.Client
}

// New creates a new MQTT Publisher
func New(cfg *config.MQTTPublisher) *Publisher {
	// connect to the MQTT broker
	options := mqtt.NewClientOptions()
	options.Servers = []*url.URL{
		{
			Scheme: cfg.Scheme,
			Host:   cfg.Host,
		},
	}
	client := mqtt.NewClient(options)
	t := client.Connect()
	_ = t.Wait()
	if t.Error() != nil {
		fmt.Println(t.Error())
		os.Exit(1)
	}

	return &Publisher{
		client: client,
	}
}

// Publish publishes a JSON payload to the configured MQTT broker
func (p *Publisher) Publish(event, button, channel string, state controller.State) {
	marshaledPayload, err := json.Marshal(payload{
		Button:  button,
		Channel: channel,
		Theme:   state.Theme,
		Colour:  state.Colour,
		Speed:   state.Speed,
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	// publish a message to the MQTT broker
	topic := fmt.Sprintf("NV7/CONTROLLER/%s", event)
	t := p.client.Publish(topic, 1, true, marshaledPayload)

	// Check for errors asynchronously
	go func() {
		_ = t.Wait()
		if t.Error() != nil {
			fmt.Println(t.Error())
		}
	}()
}
