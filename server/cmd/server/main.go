package main

import (
	"log"
	"net/http"

	"server/internal/models"
	"server/internal/service"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Warning: .env file not found, using system environment variables")
	} else {
		log.Println("✅ Loaded environment variables from .env file")
	}

	// Connect to PostgreSQL with retry
	if err := models.ConnectWithRetry(); err != nil {
		log.Fatal("Failed to connect to PostgreSQL after multiple attempts:", err)
	}

	if !models.IsConnected() {
		log.Fatal("PostgreSQL connection verification failed")
	}

	log.Println("✅ PostgreSQL connection verified!")

	router := service.NewRouter()
	log.Println("Server running on port localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", router))
}

