package config

import (
	"context"

	"github.com/sethvargo/go-envconfig"
)

// New creates config
func New() *Config {
	var cfg Config
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		panic("failed to extract default config: %s" + err.Error())
	}
	return &cfg
}

// DefaultControllerConfig returns a new default controller config
func DefaultControllerConfig() *Controller {
	var cfg Controller
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		panic("failed to extract default config: %s" + err.Error())
	}
	return &cfg
}

// Config defines all of the config required to run the controller
type Config struct {
	Controller    *Controller
	MQQTPublisher *MQTTPublisher
}

// MQTTPublisher defines the config for the MQTT Publisher
type MQTTPublisher struct {
	Scheme string `env:"MQTT_PUBLISHER_SCHEME,default=tcp"`
	Host   string `env:"MQTT_PUBLISHER_HOST,required"`
}

// Controller defines the config for the controller
type Controller struct {
	// defines the GPIO pin number used to control the channel LED on
	LEDPin int `env:"CONTROLLER_LED_PIN,default=13"`

	// define the ADS targets expected for each button
	ChannelTarget []int `env:"CONTROLLER_CHANNEL_TARGET,default=1870,1978"`
	ModeTarget    []int `env:"CONTROLLER_MODE_TARGET,default=1898,1999"`
	ColorTarget   []int `env:"CONTROLLER_COLOR_TARGET,default=1956,29"`
	SpeedTarget   []int `env:"CONTROLLER_SPEED_TARGET,default=1923,7"`

	// define a tolerance for the ADS input to allow for signal noise
	Tolerance int `env:"CONTROLLER_TOLERANCE,default=2"`

	// define the number of times a button must been seen in a poll before
	// registering a button press
	Accuracy int `env:"CONTROLLER_ACCURACY,default=2"`

	// define the number of ms required to trigger a button hold
	HoldDuration int `env:"CONTROLLER_HOLD_DURATION,default=2000"`

	// define how often the controller should check for a button press, a lower
	// number provides a more responsive input but can cause some inconsistencies
	// in the ADS reads which can lead to incorrect or missed button inputs
	PollRate int `env:"CONTROLLER_HOLD_DURATION,default=30"`
}
