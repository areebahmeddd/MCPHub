package main

import (
	"log"
	"mcphub/api"
	"mcphub/config"
)

func main() {
	// Load configuration
	cfg := config.Load()
	
	// Setup and start the API server
	router := api.SetupRouter()
	
	log.Printf("Starting MCPHub server on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
