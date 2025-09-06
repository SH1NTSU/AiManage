package models

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)





var MgC *mongo.Client

func ConnectDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	connectionString := os.Getenv("MONGO_URI")
	if connectionString == "" {
		log.Fatal("MONGO_URI not set in .env")
	}

	clientOptions := options.Client().ApplyURI(connectionString)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal("Error creating Mongo client:", err)
	}



	MgC = client
}



