package main

import (
	"log"
	"os"

	"agent-tracker/internal/database"
	"agent-tracker/internal/handlers"

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

	r := gin.Default()

	r.GET("/api/health", handlers.Health)
	r.GET("/api/tools", handlers.GetTools)
	r.GET("/api/tools/:slug", handlers.GetTool)
	r.GET("/api/tools/:slug/entries", handlers.GetToolEntries)
	r.GET("/api/entries", handlers.GetEntries)
	r.GET("/api/entries/:id", handlers.GetEntry)
	r.GET("/api/search", handlers.Search)
	r.GET("/rss/all", handlers.GetAllRSS)
	r.GET("/rss/:slug", handlers.GetToolRSS)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server starting on :" + port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}