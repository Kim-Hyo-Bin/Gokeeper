package main

import (
	"log"

	"gokeeper/internal/app"
	"gokeeper/internal/config"
)

func main() {
	if err := app.Run(config.FromEnv()); err != nil {
		log.Fatal(err)
	}
}
