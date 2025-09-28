package models

import (
	"context"
	"log"
	"go.mongodb.org/mongo-driver/mongo"

)






const DB = "AiManage"


const collectionName = "Models"

func GetCollection() *mongo.Collection {
	return MgC.Database(DB).Collection(collectionName)


}

func GetDocuments[T any](collectionName string, filter interface{}) ([]T, error) {
    collection := MgC.Database(DB).Collection(collectionName)
    cursor, err := collection.Find(context.TODO(), filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(context.TODO())

    var results []T
    if err := cursor.All(context.TODO(), &results); err != nil {
        return nil, err
    }

    log.Println("Data retrieved")
    return results, nil
}

func Insert[T any](collectionName string, document T) error {
    collection := MgC.Database(DB).Collection(collectionName)
    inserted, err := collection.InsertOne(context.TODO(), document)
    if err != nil {
        return err
    }

    log.Println("Inserted ID:", inserted.InsertedID)
    return nil
}





