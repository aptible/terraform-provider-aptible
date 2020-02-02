package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
)

func main() {
	initLogger()
	AppLogger.Println("INFO - Starting...")

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return Provider()
		},
	})
}

// CreateLogger
var AppLogger *log.Logger

func initLogger() {
	f, err := os.OpenFile("logs/create-app.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(1)
	}

	AppLogger = log.New(f, "Logger:\t", log.Ldate|log.Ltime|log.Lshortfile)
}
