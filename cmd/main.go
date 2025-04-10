package main

import (
	"log"
	"renfound_v1/internal/app"
)

func main() {
	// Create application
	application, err := app.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Run application
	if err := application.Run(); err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}
}
