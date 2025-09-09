package types

import "go.mongodb.org/mongo-driver/bson/primitive"



type Model struct {
    ID      primitive.ObjectID `bson:"_id,omitempty"`
    Name    string             `bson:"name"`
    Picture string             `bson:"picture"` // file path or URL
    Folder  []string           `bson:"folder"`  // file paths or URLs
}
