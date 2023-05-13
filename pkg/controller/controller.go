package controller

import (
	"fmt"
	"time"

	"github.com/wamphlett/nv7-pi-controller/config"
)

type button string

const (
	ButtonNone    button = "NONE"
	ButtonChannel button = "CHANNEL"
	ButtonMode    button = "MODE"
)

type targetRange struct {
	Button button
	upper  int
	lower  int
}

func (r *targetRange) InRange(input int) bool {
	return input >= r.lower && input <= r.upper
}

type Controller struct {
	previousButton button

	lastTimeIdle time.Time
	holdDuration time.Duration
	isHeld       bool

	targets []*targetRange
}

func New(cfg *config.Controller, opts ...Opt) *Controller {
	c := &Controller{}

	c.targets = configureButton(ButtonChannel, cfg.ChannelTarget, cfg.TargetRange)
	c.targets = append(c.targets, configureButton(ButtonMode, cfg.ModeTarget, cfg.TargetRange)...)

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Controller) Start() {
	for i := 0; i < 10; i++ {
		c.poll()
		time.Sleep(time.Duration(time.Millisecond * 500))
	}
}

func (c *Controller) poll() {

	pinVoltage := 500

	pollTime := time.Now()

	for _, target := range c.targets {
		if !target.InRange(pinVoltage) {
			continue
		}

		if target.Button == c.previousButton {
			if !c.isHeld && time.Since(c.lastTimeIdle) > c.holdDuration {
				c.publish(target.Button, true)
				c.isHeld = true
			}
			return
		}

		// todo toggle channel if button is channel

		c.publish(target.Button, false)
		return
	}

	c.previousButton = ButtonNone
	c.lastTimeIdle = pollTime
	c.isHeld = false
}

func (c *Controller) publish(button button, isHeld bool) {
	// todo publish to all publishers
	fmt.Printf("PUSH: %s held: %t\n", button, isHeld)
}

func configureButton(b button, targets []int, r int) []*targetRange {
	targetRanges := make([]*targetRange, len(targets))
	for i, target := range targets {
		targetRanges[i] = &targetRange{
			Button: b,
			lower:  target - r,
			upper:  target + r,
		}
	}
	return targetRanges
}
