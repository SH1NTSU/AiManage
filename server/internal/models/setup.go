package models

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MgC *mongo.Client

func ConnectDB() error {
	// Load .env file (optional)
	godotenv.Load()

	connectionString := os.Getenv("MONGO_URI")
	if connectionString == "" {
		// Use simple connection without replica set
		connectionString = "mongodb://localhost:27017"
		log.Printf("MONGO_URI not set, using default: %s", connectionString)
	}

	log.Printf("Connecting to MongoDB: %s", connectionString)

	clientOptions := options.Client().ApplyURI(connectionString)
	
	// Increase timeout and add retry logic
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	// Test the connection with a longer timeout
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer pingCancel()
	
	err = client.Ping(pingCtx, nil)
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	MgC = client
	log.Println("✅ Connected to MongoDB successfully!")
	return nil
}

// Add this function to check if DB is connected
func IsConnected() bool {
	if MgC == nil {
		return false
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err := MgC.Ping(ctx, nil)
	return err == nil
}
