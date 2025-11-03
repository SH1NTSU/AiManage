package models

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	// "go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"
)

// var MgC *mongo.Client
//
// func ConnectDB() error {
// 	// Load .env file (optional)
// 	godotenv.Load()
//
// 	connectionString := os.Getenv("MONGO_URI")
// 	if connectionString == "" {
// 		// Use simple connection without replica set
// 		connectionString = "mongodb://localhost:27017"
// 		log.Printf("MONGO_URI not set, using default: %s", connectionString)
// 	}
//
// 	log.Printf("Connecting to MongoDB: %s", connectionString)
//
// 	clientOptions := options.Client().ApplyURI(connectionString)
//
// 	// Increase timeout and add retry logic
// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()
//
// 	client, err := mongo.Connect(ctx, clientOptions)
// 	if err != nil {
// 		return fmt.Errorf("connection failed: %w", err)
// 	}
//
// 	// Test the connection with a longer timeout
// 	pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer pingCancel()
//
// 	err = client.Ping(pingCtx, nil)
// 	if err != nil {
// 		return fmt.Errorf("ping failed: %w", err)
// 	}
//
// 	MgC = client
// 	log.Println("✅ Connected to MongoDB successfully!")
// 	return nil
// }
//
// // Add this function to check if DB is connected
// func IsConnected() bool {
// 	if MgC == nil {
// 		return false
// 	}
//
// 	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
// 	defer cancel()
//
// 	err := MgC.Ping(ctx, nil)
// 	return err == nil
// }
//
// func ConnectMongoDBWithRetry() error {
// 	maxRetries := 5
// 	retryDelay := 2 * time.Second
//
// 	for i := 0; i < maxRetries; i++ {
// 		log.Printf("Attempting MongoDB connection (attempt %d/%d)...", i+1, maxRetries)
//
// 		err := ConnectDB()
// 		if err == nil {
// 			return nil // Success!
// 		}
//
// 		log.Printf("Connection failed: %v", err)
//
// 		if i < maxRetries-1 {
// 			log.Printf("Retrying in %v...", retryDelay)
// 			time.Sleep(retryDelay)
// 			retryDelay *= 2 // Exponential backoff
// 		}
// 	}
//
// 	return fmt.Errorf("failed to connect after %d attempts", maxRetries)
// }


var Pool *pgxpool.Pool

func Connect() error {
	godotenv.Load()

	dsn := os.Getenv("DB_URI")
	if dsn == "" {
		return fmt.Errorf("DB_URI not set in environment")
	}

	log.Printf("Connecting to PostgreSQL...")

	// Create connection pool with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	// Test the connection with ping
	pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer pingCancel()

	err = pool.Ping(pingCtx)
	if err != nil {
		pool.Close()
		return fmt.Errorf("ping failed: %w", err)
	}

	Pool = pool
	log.Println("✅ Connected to PostgreSQL successfully!")
	return nil
}

// IsConnected checks if the PostgreSQL connection pool is still active
func IsConnected() bool {
	if Pool == nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := Pool.Ping(ctx)
	return err == nil
}

// ConnectWithRetry attempts to connect to PostgreSQL with retry logic
func ConnectWithRetry() error {
	maxRetries := 5
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting PostgreSQL connection (attempt %d/%d)...", i+1, maxRetries)

		err := Connect()
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
