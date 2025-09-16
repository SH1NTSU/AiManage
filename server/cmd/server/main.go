package main

import (
	"fmt"
	"log"
	"net/http"
	"server/internal/models"
	"server/internal/service"
	"time"
)

func main() {
	// Initialize MongoDB with retry logic
	if err := connectMongoDBWithRetry(); err != nil {
		log.Fatal("Failed to connect to MongoDB after multiple attempts:", err)
	}

	// Verify connection
	if !models.IsConnected() {
		log.Fatal("MongoDB connection verification failed")
	}

	log.Println("✅ MongoDB connection verified!")

	router := service.NewRouter()
	log.Println("Server running on port :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func connectMongoDBWithRetry() error {
	maxRetries := 5
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting MongoDB connection (attempt %d/%d)...", i+1, maxRetries)
		
		err := models.ConnectDB()
		if err == nil {
			return nil // Success!
		}

		log.Printf("Connection failed: %v", err)
		
		if i < maxRetries-1 {
			log.Printf("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}
	}

	return fmt.Errorf("failed to connect after %d attempts", maxRetries)
}
