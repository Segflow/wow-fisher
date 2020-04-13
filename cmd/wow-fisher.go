package main

import (
	"log"

	"github.com/segflow/wow-fisher/pkg/config"
	"github.com/segflow/wow-fisher/pkg/robot"
)

func main() {
	cfg, err := config.ReadConfig("./bot.json")
	if err != nil {
		log.Fatal(err)
	}

	bot, err := robot.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	bot.Start()
}
