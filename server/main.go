package main

import (
	"log"
	"os"

	"agent-tracker/internal/database"
	"agent-tracker/internal/handlers"
	"agent-tracker/internal/sync"

	"github.com/gin-gonic/gin"
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

	if err := sync.EnsureSeeded(); err != nil {
		log.Printf("Initial sync failed: %v", err)
	}

	r := gin.Default()

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

	port := os.Getenv("PORT")
	if port == "" {
		port = "10001"
	}

	log.Println("Server starting on :" + port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
