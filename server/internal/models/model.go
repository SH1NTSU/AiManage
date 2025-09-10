package models

import (
	"context"
	"log"
	"server/internal/types"
	"go.mongodb.org/mongo-driver/mongo"

)






const DB = "AiManage"

const CollectionName = "Models"


func GetCollection() *mongo.Collection {
	return MgC.Database(DB).Collection(CollectionName)


}



func GetModels(filter any) ([]types.Model, error) {
    collection := MgC.Database(DB).Collection(CollectionName)
    cursor, err := collection.Find(context.TODO(), filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(context.TODO())

    var models []types.Model
    if err := cursor.All(context.TODO(), &models); err != nil {
        return nil, err
    }

    log.Println("Data retrieved")
    return models, nil
}

func Insert(model types.Model) error {
	

	collection := MgC.Database(DB).Collection(CollectionName)
	inserted, err := collection.InsertOne(context.TODO(), model)
	if err != nil {
		panic(err)
	}

	log.Println("Model inserted: ", inserted.InsertedID)
	
	return err


}






