// types/model.go
package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type Model struct {
    ID      primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
    Name    string             `bson:"name" json:"name"`
    Picture string             `bson:"picture" json:"picture"`
    Folder  []string           `bson:"folder" json:"folder"`
}
