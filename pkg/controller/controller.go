package controller

import (
	"fmt"
	"os"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
	"github.com/wamphlett/nv7-pi-controller/config"
)

type event string

const (
	EventButtonPress      event = "BUTTON_PRESS"
	EventButtonHold       event = "BUTTON_HOLD"
	EventChannelTurnedOff event = "CHANNEL_TURNED_OFF"
	EventChannelTurnedOn  event = "CHANNEL_TURNED_ON"
	EventStart            event = "START"
)

type channel string

const (
	ChannelA channel = "A"
	ChannelB channel = "B"
)

type button string

const (
	ButtonNone    button = "NONE"
	ButtonChannel button = "CHANNEL"
	ButtonMode    button = "MODE"
	ButtonSpeed   button = "SPEED"
	ButtonColour  button = "COLOUR"
)

type buttonRegister struct {
	registerTime time.Time
	button       button
	accuracy     int
	held         bool
}

type targetRange struct {
	Button button
	upper  int
	lower  int
}

func (r *targetRange) InRange(input int) bool {
	return input >= r.lower && input <= r.upper
}

type Controller struct {
	sampler *Sampler

	holdDuration time.Duration
	pollRate     time.Duration

	buttonRegister *buttonRegister
	currentChannel channel
	targets        []*targetRange

	ledPin rpio.Pin

	close chan (struct{})

	channelState map[channel]bool
	speed        int
	themeIndex   map[channel]int
	themes       map[channel][]theme
}

func New(cfg *config.Controller, opts ...Opt) *Controller {
	c := &Controller{
		holdDuration: time.Second * 2,
		pollRate:     time.Millisecond * 30,
		close:        make(chan struct{}),
		themeIndex: map[channel]int{
			ChannelA: 0,
			ChannelB: 0,
		},
		channelState: map[channel]bool{
			ChannelA: true,
			ChannelB: true,
		},
	}

	c.targets = configureButton(ButtonChannel, cfg.ChannelTarget, cfg.Tolerance)
	c.targets = append(c.targets, configureButton(ButtonMode, cfg.ModeTarget, cfg.Tolerance)...)
	c.targets = append(c.targets, configureButton(ButtonColour, cfg.ColorTarget, cfg.Tolerance)...)
	c.targets = append(c.targets, configureButton(ButtonSpeed, cfg.SpeedTarget, cfg.Tolerance)...)

	// Test themes
	c.themes = map[channel][]theme{
		ChannelA: {
			{
				name:    "Theme 1",
				colours: []string{"red", "green", "blue"},
			},
			{
				name:    "Theme 2",
				colours: []string{"red", "green", "blue"},
			},
		},
		ChannelB: {
			{
				name:    "Theme 3",
				colours: []string{"red", "green", "blue"},
			},
			{
				name:    "Theme 4",
				colours: []string{"red", "green", "blue"},
			},
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	// start a sampler
	c.sampler = NewSampler()

	// LED
	c.ledPin = rpio.Pin(13)

	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	c.ledPin.Output()

	// set the current channel to A on start up
	c.setChannel(ChannelA)

	c.publish(EventStart, ButtonNone)

	return c
}

func (c *Controller) Start() {
	ticker := time.NewTicker(c.pollRate)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.poll()
			case <-c.close:
				ticker.Stop()
				return
			}
		}
	}()
}

func (c *Controller) Shutdown() {
	// stop the sampler
	c.sampler.Stop()
	c.close <- struct{}{}
	rpio.Close()
}

func (c *Controller) poll() {
	pollTime := time.Now()
	result := c.sampler.Read()

	for _, target := range c.targets {
		if !target.InRange(int(result)) {
			continue
		}

		if c.buttonRegister == nil || c.buttonRegister.button != target.Button {
			c.buttonRegister = &buttonRegister{
				registerTime: pollTime,
				button:       target.Button,
			}
		}
		// increase the accuracy
		c.buttonRegister.accuracy++

		if c.buttonRegister.accuracy == 2 {
			c.handlePress(target.Button)
		}

		// if the button was the same as the previous poll, check if its being held
		if !c.buttonRegister.held && time.Since(c.buttonRegister.registerTime) > c.holdDuration {
			c.handleHold(target.Button)
			c.buttonRegister.held = true
		}

		return
	}

	// if we haven't matched a button, set everything back to idle
	c.buttonRegister = nil
}

func (c *Controller) nextSpeed() {
	c.speed = (c.speed + 1) % 3
}

func (c *Controller) nextTheme() {
	c.themeIndex[c.currentChannel] = (c.themeIndex[c.currentChannel] + 1) % len(c.themes[c.currentChannel])
}

func (c *Controller) nextColour() {
	c.themes[c.currentChannel][c.themeIndex[c.currentChannel]].nextColour()
}

func (c *Controller) turnOffChannel() {
	c.channelState[c.currentChannel] = false
}

func (c *Controller) turnOnChannel() {
	c.channelState[c.currentChannel] = true
}

func (c *Controller) handlePress(button button) {
	switch button {
	case ButtonChannel:
		c.toggleChannel()
	case ButtonMode:
		// if the channel is not on, turn it on before changing the theme
		if !c.channelState[c.currentChannel] {
			c.turnOnChannel()
			c.publish(EventChannelTurnedOn, button)
			break
		}
		c.nextTheme()
	case ButtonSpeed:
		c.nextSpeed()
	case ButtonColour:
		c.nextColour()
	}

	c.publish(EventButtonPress, button)
}

func (c *Controller) handleHold(button button) {
	switch button {
	case ButtonMode:
		c.turnOffChannel()
		c.publish(EventChannelTurnedOff, button)
	}

	c.publish(EventButtonHold, button)
}

func (c *Controller) toggleChannel() {
	if c.currentChannel == ChannelA {
		c.setChannel(ChannelB)
		return
	}
	c.setChannel(ChannelA)
}

func (c *Controller) setChannel(channel channel) {
	if channel == ChannelA {
		c.ledPin.Low()
	} else {
		c.ledPin.High()
	}
	c.currentChannel = channel
}

func (c *Controller) publish(event event, button button) {
	// todo publish to all publishers
	currentTheme := c.themes[c.currentChannel][c.themeIndex[c.currentChannel]]
	fmt.Printf("Channel: %s. event: %s. button: %s. state: %v\n", c.currentChannel, event, button, State{
		Speed:  c.speed + 1,
		Theme:  currentTheme.name,
		Colour: currentTheme.colours[currentTheme.colourIndex],
	})
}

func configureButton(b button, targets []int, tolerance int) []*targetRange {
	targetRanges := make([]*targetRange, len(targets))
	for i, target := range targets {
		targetRanges[i] = &targetRange{
			Button: b,
			lower:  target - tolerance,
			upper:  target + tolerance,
		}
	}
	return targetRanges
}
