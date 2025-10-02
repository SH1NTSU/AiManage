// // service/websocket.go
// package service
//
// import (
// 	"context"
// 	"log"
// 	"net/http"
// 	"server/internal/models"
// 	"server/internal/types"
// 	"time"
//
// 	"github.com/gorilla/websocket"
// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
// 	"go.mongodb.org/mongo-driver/mongo/options"
// )
//
// var Upgrader = websocket.Upgrader{
// 	CheckOrigin: func(r *http.Request) bool {
// 		return true
// 	},
// }
//
// func WsHandler(w http.ResponseWriter, r *http.Request) {
// 	conn, err := Upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Println("Error upgrading: ", err)
// 		return
// 	}
// 	defer conn.Close()
// 	log.Println("WebSocket client connected:", r.RemoteAddr)
//
// 	// Send initial data
// 	sendCurrentModels(conn)
//
// 	// Try change streams
// 	if err := useChangeStreams(conn); err != nil {
// 		log.Println("Change stream failed, falling back to polling:", err)
// 		usePolling(conn)
// 	}
// }
//
// func useChangeStreams(conn *websocket.Conn) error {
// 	collection := models.GetCollection()
// 	ctx := context.Background()
//
// 	stream, err := collection.Watch(ctx, mongo.Pipeline{}, options.ChangeStream().SetFullDocument(options.UpdateLookup))
// 	if err != nil {
// 		return err
// 	}
// 	defer stream.Close(ctx)
//
// 	log.Println("Change stream started successfully!")
//
// 	for stream.Next(ctx) {
// 		var changeEvent struct {
// 			OperationType string      `bson:"operationType"`
// 			FullDocument  interface{} `bson:"fullDocument"`
// 		}
//
// 		if err := stream.Decode(&changeEvent); err != nil {
// 			log.Println("Error decoding change event:", err)
// 			continue
// 		}
//
// 		log.Printf("Database change detected: %s", changeEvent.OperationType)
//
// 		if err := sendCurrentModels(conn); err != nil {
// 			log.Println("Error sending models update:", err)
// 			return err // exit change stream on write error
// 		}
// 	}
//
// 	if err := stream.Err(); err != nil {
// 		return err
// 	}
//
// 	return nil
// }
//
// func usePolling(conn *websocket.Conn) {
// 	ticker := time.NewTicker(5 * time.Second)
// 	defer ticker.Stop()
//
// 	for {
// 		select {
// 		case <-ticker.C:
// 			if err := sendCurrentModels(conn); err != nil {
// 				log.Println("Polling send error:", err)
// 				return
// 			}
// 		}
// 	}
// }
//
// func sendCurrentModels(conn *websocket.Conn) error {
// 	allModels, err := models.GetDocuments[types.Model]("Models", bson.M{})
// 	if err != nil {
// 		log.Println("GetDocuments error:", err)
// 		return err
// 	}
//
// 	if allModels == nil {
// 		allModels = []types.Model{}
// 	}
//
// 	if err := conn.WriteJSON(allModels); err != nil {
// 		log.Println("WebSocket send error:", err)
// 		return err
// 	}
//
// 	log.Println("Sent models update to client")
// 	return nil
// }



// service/websocket.go
package service

import (
    "context"
    "log"
    "net/http"
    "server/internal/models"
    "server/internal/types"
    "sync"

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

// Global variables for managing change streams
var (
    changeStreamMutex sync.Mutex
    globalChangeStream *mongo.ChangeStream
    clients = make(map[*websocket.Conn]bool)
    broadcast = make(chan []types.Model)
)

func WsHandler(w http.ResponseWriter, r *http.Request) {
    conn, err := Upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Error upgrading: ", err)
        return
    }
    defer conn.Close()
    
    log.Println("WebSocket client connected:", r.RemoteAddr)

    // Register client
    changeStreamMutex.Lock()
    clients[conn] = true
    changeStreamMutex.Unlock()

    // Send initial data
    if err := sendCurrentModels(conn); err != nil {
        log.Println("Error sending initial models:", err)
        return
    }

    // Start change stream manager (only once)
    changeStreamMutex.Lock()
    if globalChangeStream == nil {
        go manageChangeStream()
    }
    changeStreamMutex.Unlock()

    // Keep connection alive and handle client messages
    for {
        // Read messages from client (or just check if connection is alive)
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            log.Println("WebSocket read error:", err)
            break
        }
        
        // Handle ping/pong or other messages
        if messageType == websocket.PingMessage {
            if err := conn.WriteMessage(websocket.PongMessage, nil); err != nil {
                log.Println("WebSocket pong error:", err)
                break
            }
        }
        
        log.Printf("Received message: %s", p)
    }

    // Unregister client
    changeStreamMutex.Lock()
    delete(clients, conn)
    
    // Clean up change stream if no clients left
    if len(clients) == 0 && globalChangeStream != nil {
        globalChangeStream.Close(context.Background())
        globalChangeStream = nil
    }
    changeStreamMutex.Unlock()
    
    log.Println("WebSocket client disconnected:", r.RemoteAddr)
}

func manageChangeStream() {
    ctx := context.Background()
    collection := models.GetCollection()
    
    var err error
    changeStreamMutex.Lock()
    globalChangeStream, err = collection.Watch(ctx, mongo.Pipeline{}, 
        options.ChangeStream().SetFullDocument(options.UpdateLookup))
    changeStreamMutex.Unlock()
    
    if err != nil {
        log.Println("Change stream failed:", err)
        return
    }
    
    defer func() {
        changeStreamMutex.Lock()
        if globalChangeStream != nil {
            globalChangeStream.Close(ctx)
            globalChangeStream = nil
        }
        changeStreamMutex.Unlock()
    }()

    log.Println("Global change stream started successfully!")

    for globalChangeStream.Next(ctx) {
        var changeEvent struct {
            OperationType string      `bson:"operationType"`
            FullDocument  interface{} `bson:"fullDocument"`
        }

        if err := globalChangeStream.Decode(&changeEvent); err != nil {
            log.Println("Error decoding change event:", err)
            continue
        }

        log.Printf("Database change detected: %s", changeEvent.OperationType)

        // Get updated models and broadcast to all clients
        allModels, err := models.GetDocuments[types.Model]("Models", bson.M{})
        if err != nil {
            log.Println("GetDocuments error:", err)
            continue
        }

        if allModels == nil {
            allModels = []types.Model{}
        }

        // Broadcast to all connected clients
        changeStreamMutex.Lock()
        for client := range clients {
            if err := client.WriteJSON(allModels); err != nil {
                log.Println("Error broadcasting to client:", err)
                client.Close()
                delete(clients, client)
            }
        }
        changeStreamMutex.Unlock()
        
        log.Printf("Broadcasted models update to %d clients", len(clients))
    }

    if err := globalChangeStream.Err(); err != nil {
        log.Println("Change stream error:", err)
    }
}

func sendCurrentModels(conn *websocket.Conn) error {
    allModels, err := models.GetDocuments[types.Model]("Models", bson.M{})
    if err != nil {
        log.Println("GetDocuments error:", err)
        return err
    }

    if allModels == nil {
        allModels = []types.Model{}
    }

    if err := conn.WriteJSON(allModels); err != nil {
        log.Println("WebSocket send error:", err)
        return err
    }

    log.Println("Sent models update to client")
    return nil
}
