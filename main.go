package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/wamphlett/nv7-pi-controller/config"
	"github.com/wamphlett/nv7-pi-controller/pkg/controller"
	"github.com/wamphlett/nv7-pi-controller/pkg/mqtt"
)

func main() {

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	cfg := config.New()

	fmt.Println(cfg.Controller)

	mqqtPublisher := mqtt.New("")
	c := controller.New(cfg.Controller, controller.WithPublisher(mqqtPublisher))
	c.Start()

	// wait for shutdown
	<-signals

	c.Shutdown()
}
