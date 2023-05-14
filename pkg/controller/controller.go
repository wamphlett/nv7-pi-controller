package controller

import (
	"fmt"
	"os"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
	"github.com/wamphlett/nv7-pi-controller/config"
)

// event defines the events
type event string

const (
	EventButtonPress      event = "BUTTON_PRESS"
	EventButtonHold       event = "BUTTON_HOLD"
	EventChannelTurnedOff event = "CHANNEL_TURNED_OFF"
	EventChannelTurnedOn  event = "CHANNEL_TURNED_ON"
	EventStart            event = "START"
)

// channel defines the available channels on the controller
type channel string

const (
	ChannelA channel = "A"
	ChannelB channel = "B"
)

// button defines all of the configure buttons
type button string

const (
	ButtonNone    button = "NONE"
	ButtonChannel button = "CHANNEL"
	ButtonMode    button = "MODE"
	ButtonSpeed   button = "SPEED"
	ButtonColour  button = "COLOUR"
)

// Publisher defines the methods required for publishers
type Publisher interface {
	Publish(event, button, channel string, state State)
}

// Sampler defines the methods required of the sampler
type Sampler interface {
	Read() float64
}

// buttonRegister is responsile for keeping a history registered buttons
type buttonRegister struct {
	registerTime time.Time
	button       button
	accuracy     int
	held         bool
}

// targetRange specifies an upper and lower ADS reading and defines
// which button that range belongs to
type targetRange struct {
	Button button
	upper  int
	lower  int
}

// InRange will return true when the given ADS reading is between the
// upper and lower limit of the target
func (r *targetRange) InRange(input int) bool {
	return input >= r.lower && input <= r.upper
}

// Controller defines the controller
type Controller struct {
	sampler Sampler

	holdDuration   time.Duration
	pollRate       time.Duration
	accuracyTarget int

	ledPin rpio.Pin

	buttonRegister *buttonRegister
	currentChannel channel
	targets        []*targetRange

	channelState map[channel]bool
	speed        int
	themeIndex   map[channel]int
	themes       map[channel][]theme

	publishers []Publisher

	close chan (struct{})
}

// New creates a new fully configured controller
func New(cfg *config.Controller, sampler Sampler, opts ...Opt) *Controller {
	c := &Controller{
		holdDuration:   time.Millisecond * time.Duration(cfg.HoldDuration),
		pollRate:       time.Millisecond * time.Duration(cfg.PollRate),
		accuracyTarget: cfg.Accuracy,
		sampler:        sampler,
		close:          make(chan struct{}),
		themeIndex: map[channel]int{
			ChannelA: 0,
			ChannelB: 0,
		},
		channelState: map[channel]bool{
			ChannelA: true,
			ChannelB: true,
		},
		publishers: []Publisher{},
	}

	// configure the ADS targets for each button
	c.targets = configureTargets(ButtonChannel, cfg.ChannelTarget, cfg.Tolerance)
	c.targets = append(c.targets, configureTargets(ButtonMode, cfg.ModeTarget, cfg.Tolerance)...)
	c.targets = append(c.targets, configureTargets(ButtonColour, cfg.ColorTarget, cfg.Tolerance)...)
	c.targets = append(c.targets, configureTargets(ButtonSpeed, cfg.SpeedTarget, cfg.Tolerance)...)

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

	// apply controller options
	for _, opt := range opts {
		opt(c)
	}

	// configure the GPIO pin for the LED
	c.ledPin = rpio.Pin(cfg.LEDPin)
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	c.ledPin.Output()

	// set the current channel to A on start up and publish
	// a start up event
	c.setChannel(ChannelA)
	c.publish(EventStart, ButtonNone)

	return c
}

// Start starts the controller polling
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

// Shutdown cleanly exists the controller
func (c *Controller) Shutdown() {
	c.close <- struct{}{}
	rpio.Close()
}

// poll reads the ADS value and checks against the configured target
func (c *Controller) poll() {
	pollTime := time.Now()
	result := c.sampler.Read()

	for _, target := range c.targets {
		if !target.InRange(int(result)) {
			continue
		}

		// check if a button has been registered or register a new one
		if c.buttonRegister == nil || c.buttonRegister.button != target.Button {
			c.buttonRegister = &buttonRegister{
				registerTime: pollTime,
				button:       target.Button,
			}
		}

		// each time we see the button, increase the accuracy.
		// this prevents incorrect buttons being selected due to noise from the ADS reader
		c.buttonRegister.accuracy++

		// only register a button press when the desired accuracy has been hit
		if c.buttonRegister.accuracy == c.accuracyTarget {
			c.handlePress(target.Button)
		}

		// if the button was the same as the previous poll, check if its being held and register a
		// button hold
		if !c.buttonRegister.held && time.Since(c.buttonRegister.registerTime) > c.holdDuration {
			c.handleHold(target.Button)
			c.buttonRegister.held = true
		}

		return
	}

	// if we haven't matched a button, clear the button register
	c.buttonRegister = nil
}

// nextSpeed increments the speed
func (c *Controller) nextSpeed() {
	c.speed = (c.speed + 1) % 3
}

// nextTheme increments the next theme for the current channel
func (c *Controller) nextTheme() {
	c.themeIndex[c.currentChannel] = (c.themeIndex[c.currentChannel] + 1) % len(c.themes[c.currentChannel])
}

// nextColour increments the next colour for the current channel and theme
func (c *Controller) nextColour() {
	c.themes[c.currentChannel][c.themeIndex[c.currentChannel]].nextColour()
}

// turnOffChannel registers the current channel as off
func (c *Controller) turnOffChannel() {
	c.channelState[c.currentChannel] = false
}

// turnOffChannel registers the current channel as on
func (c *Controller) turnOnChannel() {
	c.channelState[c.currentChannel] = true
}

// handlePress handles button presses
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

// handleHold handles button holds
func (c *Controller) handleHold(button button) {
	switch button {
	case ButtonMode:
		c.turnOffChannel()
		c.publish(EventChannelTurnedOff, button)
	}

	c.publish(EventButtonHold, button)
}

// toggleChannel changes the channel
func (c *Controller) toggleChannel() {
	if c.currentChannel == ChannelA {
		c.setChannel(ChannelB)
		return
	}
	c.setChannel(ChannelA)
}

// setChannel sets the channel and updates the LED
func (c *Controller) setChannel(channel channel) {
	if channel == ChannelA {
		c.ledPin.Low()
	} else {
		c.ledPin.High()
	}
	c.currentChannel = channel
}

// publish publishes the event to all of the configured publishers
func (c *Controller) publish(event event, button button) {
	currentTheme := c.themes[c.currentChannel][c.themeIndex[c.currentChannel]]
	state := State{
		Speed:  c.speed + 1,
		Theme:  currentTheme.name,
		Colour: currentTheme.colours[currentTheme.colourIndex],
	}
	fmt.Printf("Channel: %s. event: %s. button: %s. state: %v\n", c.currentChannel, event, button, state)

	// send the event to all publishers
	for _, publisher := range c.publishers {
		publisher.Publish(string(event), string(button), string(c.currentChannel), state)
	}
}

// configureTargets sets the target ranges for the given button
func configureTargets(b button, targets []int, tolerance int) []*targetRange {
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
