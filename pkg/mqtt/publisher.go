package mqtt

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"github.com/wamphlett/nv7-pi-controller/pkg/controller"
)

type payload struct {
	Button  string
	Channel string
	Theme   string
	Colour  string
	Speed   int
}

type Publisher struct {
	client mqtt.Client
}

func New(host string) *Publisher {
	options := mqtt.NewClientOptions()
	options.Servers = []*url.URL{
		{
			Host: host,
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
	}
	topic := fmt.Sprintf("NV7/CONTROLLER/%s", event)
	t := p.client.Publish(topic, 1, true, marshaledPayload)
	go func() {
		_ = t.Wait()
		if t.Error() != nil {
			fmt.Println(t.Error())
		}
	}()
}
