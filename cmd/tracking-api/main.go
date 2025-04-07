package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"UTM_Tracker/internal/database"
	"UTM_Tracker/internal/handlers"
	"UTM_Tracker/internal/tracking"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database
	db := database.NewPostgresDatabase()

	// Initialize click tracker
	clickTracker := tracking.NewClickTracker(db.Client)

	// Create redirect handler
	redirectHandler := handlers.NewRedirectHandler(db, clickTracker)

	// Setup Gin router
	router := gin.Default()

	// Middleware for CORS and security
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// Redirect route
	router.GET("/r/:slug", redirectHandler.HandleRedirect)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	log.Fatal(router.Run(":" + port))
}
