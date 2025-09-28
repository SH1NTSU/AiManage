// types/model.go
package types

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Model struct {
    ID      primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
    Name    string             `bson:"name" json:"name"`
    Picture string             `bson:"picture" json:"picture"`
    Folder  []string           `bson:"folder" json:"folder"`
}



type User struct {
    ID      primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
    Email string `bson:"email" json:"email"`
    Password string `bson:"password" json:"password"`
}


type Session struct {
	
    ID      primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty"`
    
    Email string `bson:"email" json:"email"`
    Refresh_token string `bson:"refresh_token" json:"refresh_token"`
    Expires_at time.Time `bson:"expires_at" json:"expires_at"`
}
