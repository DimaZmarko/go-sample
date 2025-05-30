package main

import (
	"log"
	"go-sample/internal/app"
)

func main() {
	// Initialize and start the application
	application, err := app.NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Start the server
	if err := application.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
} 