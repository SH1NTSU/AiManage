package models

import (
	"context"
	"server/internal/types"
	"log"

)








const DB = "AiManage"

const collectionName = "Models"

func Insert(model types.Model) error {
	

	collection := MgC.Database(DB).Collection(collectionName)
	inserted, err := collection.InsertOne(context.TODO(), model)
	if err != nil {
		panic(err)
	}

	log.Println("Model inserted: ", inserted.InsertedID)
	
	return err


}
