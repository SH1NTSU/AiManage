package service

import (
	"net/http"
	"server/internal/handlers"
	"server/internal/middlewares"

	"github.com/go-chi/chi/v5"
)



func NewRouter() http.Handler {
    r := chi.NewRouter()
    
	r.Use(middlewares.WithCORS)	
	r.Route("/v1", func(r chi.Router) {
		
		
		r.HandleFunc("/ws", WsHandler)
		
		r.Post("/register", handlers.RegisterHandler)
		r.Post("/login", handlers.LoginHandler)
		r.Get("/refresh", handlers.RefreshHandler)
		r.Group(func(protected chi.Router) {
			protected.Use(middlewares.JWTGuard)
			protected.Get("/health", handlers.HealthCheckHandler)

			protected.Post("/insert", handlers.InsertHandler)
			protected.Get("/getModels", handlers.ReadHandler)
			protected.Delete("/deleteModel", handlers.DeleteModel)
		})
	})	
	return r


}






