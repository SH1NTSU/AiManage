// service/websocket.go
package service

import (
	"context"
	"log"
	"net/http"
	"server/internal/models"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading: ", err)
		return
	}
	defer conn.Close()

	// Send initial data immediately
	sendCurrentModels(conn)

	// Try to use change streams first
	if useChangeStreams(conn) {
		log.Println("WebSocket client connected using change streams")
		return
	}

	// Fallback to polling if change streams fail
	log.Println("WebSocket client connected, falling back to polling...")
	usePolling(conn)
}

func useChangeStreams(conn *websocket.Conn) bool {
	collection := models.GetCollection()
	ctx := context.Background()

	// Create change stream with full document lookup
	pipeline := mongo.Pipeline{}
	opts := options.ChangeStream().SetFullDocument(options.UpdateLookup)

	stream, err := collection.Watch(ctx, pipeline, opts)
	if err != nil {
		log.Println("Change stream error, falling back to polling:", err)
		return false
	}
	defer stream.Close(ctx)

	log.Println("Change stream started successfully!")

	// Process change events in real-time
	for stream.Next(ctx) {
		var changeEvent struct {
			OperationType string      `bson:"operationType"`
			FullDocument  interface{} `bson:"fullDocument"`
		}
		
		if err := stream.Decode(&changeEvent); err != nil {
			log.Println("Error decoding change event:", err)
			continue
		}

		log.Printf("Database change detected: %s", changeEvent.OperationType)
		
		// Send updated models to client
		sendCurrentModels(conn)
	}

	if err := stream.Err(); err != nil {
		log.Println("Change stream error:", err)
		return false
	}

	return true
}

func usePolling(conn *websocket.Conn) {
	ticker := time.NewTicker(5 * time.Second) // Longer interval for fallback
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sendCurrentModels(conn)
		}
	}
}

func sendCurrentModels(conn *websocket.Conn) {
	allModels, err := models.GetModels(bson.M{})
	if err != nil {
		log.Println("GetModels error: ", err)
		return
	}
	if err := conn.WriteJSON(allModels); err != nil {
		log.Println("websocket send error: ", err)
		return
	}
}
