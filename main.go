package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/wamphlett/nv7-pi-controller/config"
	"github.com/wamphlett/nv7-pi-controller/pkg/controller"
	"github.com/wamphlett/nv7-pi-controller/pkg/mqtt"
	"github.com/wamphlett/nv7-pi-controller/pkg/sampler"
)

func main() {
	// set up handlers to cleanly shutdown the program
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	cfg := config.New()

	mqqtPublisher := mqtt.New(cfg.MQQTPublisher)

	s := sampler.New()
	c := controller.New(cfg.Controller, s, controller.WithPublisher(mqqtPublisher))
	s.Start()
	c.Start()

	// wait for shutdown
	<-signals

	s.Stop()
	c.Shutdown()
}
