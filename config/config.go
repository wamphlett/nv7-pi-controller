package config

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

func New() *Config {
	var cfg Config
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		panic("failed to extract default config: %s" + err.Error())
	}
	return &cfg
}

func DefaultControllerConfig() *Controller {
	var cfg Controller
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		panic("failed to extract default config: %s" + err.Error())
	}
	return &cfg
}

type Config struct {
	Controller *Controller
}

type Controller struct {
	LEDPin int `env:"CONTROLLER_LED_PIN,default=13"`

	ChannelTarget []int `env:"CONTROLLER_CHANNEL_TARGET,default=61280"`
	ModeTarget    []int `env:"CONTROLLER_MODE_TARGET,default=62180"`
	ColorTarget   []int `env:"CONTROLLER_COLOR_TARGET,default=6400"`
	SpeedTarget   []int `env:"CONTROLLER_SPEED_TARGET,default=62950"`

	Tolerance int `env:"CONTROLLER_TOLERANCE,default=40"`
}
