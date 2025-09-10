package service

import (
	"net/http"
    "github.com/go-chi/chi/v5"
	"server/internal/handlers"
	"github.com/go-chi/cors"

)


func NewRouter() http.Handler { 



	r := chi.NewRouter()
	

	
	r.Use(cors.Handler(cors.Options{
	    AllowedOrigins:   []string{"http://localhost:3000"},
	    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
	    AllowCredentials: true,
	}))
	
	
	r.Get("/health", handlers.HealthCheckHandler)
	
	r.Post("/api/v1/insert", handlers.InsertHandler)

	r.Get("/api/v1/getModels", handlers.ReadHandler)
	
	r.HandleFunc("/ws", WsHandler)
	return r


}






