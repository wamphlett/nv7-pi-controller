package main

import (
	"fmt"
	"math"
	"os"
	"os/signal"
	"syscall"

	"github.com/wamphlett/nv7-pi-controller/config"
	"github.com/wamphlett/nv7-pi-controller/pkg/controller"
)

func main() {

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	cfg := config.New()

	fmt.Println(cfg.Controller)

	c := controller.New(cfg.Controller)
	c.Start()

	// wait for shutdown
	<-signals

	c.Shutdown()
}
