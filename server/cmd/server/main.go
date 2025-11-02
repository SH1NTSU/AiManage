package main

import (
	"log"
	"net/http"
	"server/internal/models"
	"server/internal/service"
)

func main() {
	if err := models.ConnectMongoDBWithRetry(); err != nil {
		log.Fatal("Failed to connect to MongoDB after multiple attempts:", err)
	}

	if !models.IsConnected() {
		log.Fatal("MongoDB connection verification failed")
	}

	log.Println("✅ MongoDB connection verified!")

	router := service.NewRouter()
	log.Println("Server running on port localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", router))
}

