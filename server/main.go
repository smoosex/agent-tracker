package main

import (
	"log"
	"os"

	"agent-tracker/internal/config"
	"agent-tracker/internal/database"
	"agent-tracker/internal/handlers"
	"agent-tracker/internal/sync"

	"github.com/gin-gonic/gin"
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

	if err := sync.EnsureSeeded(); err != nil {
		log.Printf("Initial sync failed: %v", err)
	}

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
	r.POST("/api/sync", handlers.TriggerSync)
	r.GET("/rss/all", handlers.GetAllRSS)
	r.GET("/rss/:slug", handlers.GetToolRSS)

	log.Println("Server starting on :" + cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
