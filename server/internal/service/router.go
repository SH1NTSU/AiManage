package service

import (
	"net/http"
	"server/internal/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)



func NewRouter() http.Handler {
    r := chi.NewRouter()
    
    r.Use(cors.Handler(cors.Options{
        AllowedOrigins:   []string{"http://localhost:5173"}, // Your Vite frontend
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Content-Length"},
        AllowCredentials: true,
        MaxAge:           300,
    }))
	
	
	r.Get("/health", handlers.HealthCheckHandler)
	
	r.Post("/api/v1/insert", handlers.InsertHandler)

	r.Get("/api/v1/getModels", handlers.ReadHandler)
	
	r.HandleFunc("/ws", WsHandler)
	return r


}






