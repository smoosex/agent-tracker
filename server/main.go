package main

import (
	"log"
	"os"
	"time"

	"agent-tracker/internal/config"
	"agent-tracker/internal/database"
	"agent-tracker/internal/handlers"
	"agent-tracker/internal/logging"
	"agent-tracker/internal/sync"

	"github.com/gin-gonic/gin"
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

	if err := sync.EnsureSeeded(); err != nil {
		log.Printf("Initial sync failed: %v", err)
	}

	startSyncScheduler()

	r := gin.Default()
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Fatal("Failed to configure trusted proxies:", err)
	}

	r.GET("/api/health", handlers.Health)
	r.GET("/api/tools", handlers.GetTools)
	r.GET("/api/tools/:slug", handlers.GetTool)
	r.GET("/api/tools/:slug/entries", handlers.GetToolEntries)
	r.GET("/api/entries", handlers.GetEntries)
	r.GET("/api/entries/:id", handlers.GetEntry)
	r.GET("/api/search", handlers.Search)
	r.GET("/api/logs", handlers.GetRecentLogs)
	r.GET("/api/sync/events", handlers.GetSyncEvents)
	r.GET("/api/sync/status", handlers.GetSyncStatus)
	r.GET("/api/sync/failures", handlers.GetRecentSyncFailures)
	r.POST("/api/sync", handlers.TriggerSync)
	r.GET("/rss/all", handlers.GetAllRSS)
	r.GET("/rss/:slug", handlers.GetToolRSS)

	log.Println("Server starting on :" + cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func startSyncScheduler() {
	go func() {
		for {
			nextRun := nextScheduledSync(time.Now())
			wait := time.Until(nextRun)
			log.Printf("Next scheduled sync at %s", nextRun.Format(time.RFC3339))

			timer := time.NewTimer(wait)
			<-timer.C

			result, err := handlers.RunSync()
			if err != nil {
				log.Printf("Scheduled sync failed at %s: %v", time.Now().Format(time.RFC3339), err)
				continue
			}
			if result.HasFailures() {
				log.Printf("Scheduled sync completed with %d failures at %s: %s", result.Failed, time.Now().Format(time.RFC3339), result.FailureSummary())
				continue
			}

			log.Printf("Scheduled sync completed at %s", time.Now().Format(time.RFC3339))
		}
	}()
}

func nextScheduledSync(now time.Time) time.Time {
	location := now.Location()
	year, month, day := now.Date()
	candidates := []time.Time{
		time.Date(year, month, day, 0, 0, 0, 0, location),
		time.Date(year, month, day, 12, 0, 0, 0, location),
		time.Date(year, month, day+1, 0, 0, 0, 0, location),
	}

	for _, candidate := range candidates {
		if candidate.After(now) {
			return candidate
		}
	}

	return time.Date(year, month, day+1, 12, 0, 0, 0, location)
}
