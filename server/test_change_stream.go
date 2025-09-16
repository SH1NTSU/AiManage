package main

import (
	"context"
	"log"
	"server/internal/models"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Connect to MongoDB
	err := models.ConnectDB()
	if err != nil {
		log.Fatal("Connection failed:", err)
	}

	// Test change streams
	collection := models.GetCollection()
	ctx := context.Background()

	pipeline := mongo.Pipeline{}
	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)

	stream, err := collection.Watch(ctx, pipeline, opts)
	if err != nil {
		log.Fatal("Change stream test failed:", err)
	}

	log.Println("✅ Change streams are working!")
	stream.Close(ctx)
}
