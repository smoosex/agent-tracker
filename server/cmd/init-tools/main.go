package main

import (
	"log"
	"os"

	"agent-tracker/internal/database"
	"agent-tracker/internal/sync"
)

func main() {
	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = "./data"
	}

	dbPath := dataDir + "/agent-tracker.db"
	if err := database.Init(dbPath); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	sync.InitTools()
	log.Println("Tools initialized successfully")
}