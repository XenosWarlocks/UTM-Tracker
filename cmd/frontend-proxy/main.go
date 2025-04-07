package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type FrontendProxy struct {
	trackingAPIURL string
	client         *http.Client
}

func NewFrontendProxy() *FrontendProxy {
	return &FrontendProxy{
		trackingAPIURL: os.Getenv("TRACKING_API_URL"),
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

func (fp *FrontendProxy) HandleRedirect(c *gin.Context) {
	slug := c.Param("slug")

	// Construct tracking API request URL
	trackingURL := fmt.Sprintf("%s/r/%s", fp.trackingAPIURL, slug)

	// Forward request to tracking API
	req, err := http.NewRequest("GET", trackingURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// Copy headers from original request
	req.Header = c.Request.Header.Clone()

	// Send request to tracking API
	resp, err := fp.client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to forward request"})
		return
	}
	defer resp.Body.Close()

	// Check for redirect
	if resp.StatusCode == http.StatusTemporaryRedirect ||
		resp.StatusCode == http.StatusPermanentRedirect {
		redirectURL := resp.Header.Get("Location")
		if redirectURL == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No redirect URL found"})
			return
		}

		// Redirect client
		c.Redirect(resp.StatusCode, redirectURL)
		return
	}

	// If not a redirect, return error
	body, _ := io.ReadAll(resp.Body)
	c.String(resp.StatusCode, string(body))
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found")
	}

	// Setup Gin router
	router := gin.Default()

	// Create frontend proxy
	proxy := NewFrontendProxy()

	// Redirect route
	router.GET("/r/:slug", proxy.HandleRedirect)

	// Start server
	port := os.Getenv("FRONTEND_PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Frontend Proxy starting on port %s", port)
	log.Fatal(router.Run(":" + port))
}
