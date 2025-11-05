package service

import (
	"net/http"
	"server/aiAgent"
	"server/internal/handlers"
	"server/internal/middlewares"

	"github.com/go-chi/chi/v5"
)



func NewRouter() http.Handler {
    r := chi.NewRouter()

	r.Use(middlewares.WithCORS)

	// Serve static files from uploads directory
	fileServer := http.FileServer(http.Dir("./uploads"))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", fileServer))

	// Initialize AI Agent Handler
	aiAgentHandler, err := handlers.NewAIAgentHandler()
	if err != nil {
		// Log error but continue - AI agent is optional
		// You might want to add proper logging here
	}

	// Initialize Training Handler (if AI Agent is available)
	var trainingHandler *handlers.TrainingHandler
	if aiAgentHandler != nil {
		trainingHandler = handlers.NewTrainingHandler(aiAgentHandler.GetAgent())

		// Set up broadcast callback for training updates
		broadcaster := GetTrainingBroadcaster()
		aiAgent.SetBroadcastCallback(func(trainingID string, updateType string, data interface{}) {
			broadcaster.BroadcastTrainingUpdate(trainingID, updateType, data)
		})
	}

	r.Route("/v1", func(r chi.Router) {


		r.HandleFunc("/ws", WsHandler)
		r.HandleFunc("/ws/training", TrainingWSHandler)

		r.Post("/register", handlers.RegisterHandler)
		r.Post("/login", handlers.LoginHandler)
		r.Get("/refresh", handlers.RefreshHandler)
		r.Group(func(protected chi.Router) {
			protected.Use(middlewares.JWTGuard)
			protected.Get("/health", handlers.HealthCheckHandler)

			protected.Post("/insert", handlers.InsertHandler)
			protected.Get("/getModels", handlers.ReadHandler)
			protected.Delete("/deleteModel", handlers.DeleteModel)

			// AI Agent routes
			if aiAgentHandler != nil {
				protected.Post("/ai/analyze", aiAgentHandler.AnalyzeDirectory)
				protected.Get("/ai/directory", aiAgentHandler.GetDirectoryInfo)
				protected.Get("/ai/directories", aiAgentHandler.ListDirectories)
				protected.Post("/ai/prompt", aiAgentHandler.CustomPrompt)
			}

			// Training routes
			if trainingHandler != nil {
				protected.Post("/train/start", trainingHandler.StartTraining)
				protected.Get("/train/progress", trainingHandler.GetTrainingProgress)
				protected.Post("/train/analyze", trainingHandler.AnalyzeResults)
				protected.Post("/train/cleanup", trainingHandler.CleanupOldTrainings)
			}
		})
	})
	return r


}






