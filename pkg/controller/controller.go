package controller

import (
	"fmt"
	"os"
	"time"

	"github.com/grant-carpenter/go-ads"
	"github.com/stianeikeland/go-rpio/v4"
	"github.com/wamphlett/nv7-pi-controller/config"
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
	holdDuration time.Duration

	targets []*targetRange

	buttonKey *ads.ADS

	currentChannel channel

	ledPin rpio.Pin

	pollRate time.Duration

	close chan (struct{})

	sampler *Sampler

	buttonRegister *buttonRegister
}

func New(cfg *config.Controller, opts ...Opt) *Controller {
	c := &Controller{
		holdDuration: time.Second * 3,
		pollRate:     time.Millisecond * 30,
		close:        make(chan struct{}),
	}

	c.targets = configureButton(ButtonChannel, cfg.ChannelTarget, cfg.Tolerance)
	c.targets = append(c.targets, configureButton(ButtonMode, cfg.ModeTarget, cfg.Tolerance)...)
	c.targets = append(c.targets, configureButton(ButtonColour, cfg.ColorTarget, cfg.Tolerance)...)
	c.targets = append(c.targets, configureButton(ButtonSpeed, cfg.SpeedTarget, cfg.Tolerance)...)

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
	if err := c.buttonKey.Close(); err != nil {
		fmt.Println(err)
	}
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

func (c *Controller) handlePress(button button) {
	switch button {
	case ButtonChannel:
		c.toggleChannel()
	}

	c.publish(button, false)
}

func (c *Controller) handleHold(button button) {
	c.publish(button, true)
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

func (c *Controller) publish(button button, isHeld bool) {
	// todo publish to all publishers
	fmt.Printf("Channel %s: %s held: %t\n", c.currentChannel, button, isHeld)
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
