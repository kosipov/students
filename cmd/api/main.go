package main

import (
	"github.com/kosipov/students/config"
	"github.com/kosipov/students/server"
	"log"
	"os"
)

func main() {
	if err := config.Init(); err != nil {
		log.Fatalf("%s", err.Error())
	}

	app := server.NewApp()
	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	if err := app.Run(port); err != nil {
		log.Fatalf("%s", err.Error())
	}
}
