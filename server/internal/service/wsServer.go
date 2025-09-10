package service

import (
	"context"
	"log"
	"net/http"
	"server/internal/models"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}



func WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading: ", err)
		return
	}
	defer conn.Close()

	collection := models.GetCollection()
	ctx := context.TODO()

	// --- Send initial data immediately ---
	allModels, err := models.GetModels(bson.M{})
	if err != nil {
		log.Println("GetModels error: ", err)
	} else {
		if err := conn.WriteJSON(allModels); err != nil {
			log.Println("websocket send error: ", err)
			return
		}
	}

	// --- Start watching the change stream ---
	stream, err := collection.Watch(ctx, bson.A{})
	if err != nil {
		log.Println("Change stream error: ", err)
		return
	}
	defer stream.Close(ctx)

	log.Println("WebSocket client connected, waiting for DB changes...")

	for stream.Next(ctx) {
		allModels, err := models.GetModels(bson.M{})
		if err != nil {
			log.Println("GetModels error: ", err)
			continue
		}
		if err := conn.WriteJSON(allModels); err != nil {
			log.Println("websocket send error: ", err)
			break
		}
	}

	if err := stream.Err(); err != nil {
		log.Println("Stream error:", err)
	}
}

	


