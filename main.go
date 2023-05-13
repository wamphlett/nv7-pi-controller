package main

import (
	"fmt"

	"github.com/wamphlett/nv7-pi-controller/config"
	"github.com/wamphlett/nv7-pi-controller/pkg/controller"
)

func main() {
	cfg := config.New()

	fmt.Println(cfg.Controller)

	c := controller.New(cfg.Controller)
	c.Start()
}
