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
	result, err := sync.SyncAll(false)
	if err != nil {
		log.Fatal("Incremental sync failed:", err)
	}
	if result.HasFailures() {
		log.Printf("Incremental sync completed with %d failures: %s", result.Failed, result.FailureSummary())
		return
	}
	log.Println("Incremental sync completed")
}
