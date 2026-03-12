package main

import (
	"log"
	"os"

	"agent-tracker/internal/config"
	"agent-tracker/internal/database"
	"agent-tracker/internal/sync"
)

func main() {
	cfg, err := config.LoadFromArgs(os.Args[1:])
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	dbPath := cfg.DataDir + "/agent-tracker.db"
	if err := database.Init(dbPath); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	sync.InitTools()
	if err := sync.SyncAll(false); err != nil {
		log.Fatal("Incremental sync failed:", err)
	}
	log.Println("Incremental sync completed")
}
