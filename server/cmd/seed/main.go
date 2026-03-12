package main

import (
	"log"
	"os"

	"agent-tracker/internal/config"
	"agent-tracker/internal/database"
	"agent-tracker/internal/logging"
	"agent-tracker/internal/sync"
)

func main() {
	cfg, err := config.LoadFromArgs(os.Args[1:])
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	if err := logging.Init(cfg.LogPath); err != nil {
		log.Fatal("Failed to initialize logging:", err)
	}

	dbPath := cfg.DataDir + "/agent-tracker.db"
	if err := database.Init(dbPath); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	sync.InitTools()
	log.Println("Tools initialized, starting full sync...")

	result, err := sync.SyncAll(true)
	if err != nil {
		log.Fatal("Full seed failed:", err)
	}
	if result.HasFailures() {
		log.Printf("Full seed completed with %d failures: %s", result.Failed, result.FailureSummary())
		return
	}
	log.Println("Full seed completed")
}
