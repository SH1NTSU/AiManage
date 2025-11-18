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
	var deleteModelHandler *handlers.DeleteModelHandler
	if aiAgentHandler != nil {
		agent := aiAgentHandler.GetAgent()
		trainingHandler = handlers.NewTrainingHandler(agent)
		deleteModelHandler = handlers.NewDeleteModelHandler(agent)

		// Set global trainer for remote training support
		handlers.SetGlobalTrainer(agent.GetTrainer())

		// Set up broadcast callback for training updates
		broadcaster := GetTrainingBroadcaster()
		aiAgent.SetBroadcastCallback(func(trainingID string, updateType string, data interface{}) {
			broadcaster.BroadcastTrainingUpdate(trainingID, updateType, data)
		})
	}

	r.Route("/v1", func(r chi.Router) {


		r.HandleFunc("/ws", WsHandler)
		r.HandleFunc("/ws/training", TrainingWSHandler)
		r.HandleFunc("/ws/agent", handlers.AgentWebSocketHandler)

		// Agent model upload (uses API key auth, not JWT)
		r.Post("/agent/upload-model", handlers.UploadTrainedModelHandler)

		r.Post("/register", handlers.RegisterHandler)
		r.Post("/login", handlers.LoginHandler)
		r.Get("/refresh", handlers.RefreshHandler)

		// OAuth routes
		r.Post("/auth/google", handlers.GoogleOAuthHandler)
		r.Post("/auth/github", handlers.GitHubOAuthHandler)
		r.Post("/auth/apple", handlers.AppleOAuthHandler)
		r.Group(func(protected chi.Router) {
			protected.Use(middlewares.JWTGuard)
			protected.Get("/health", handlers.HealthCheckHandler)
			protected.Get("/me", handlers.GetCurrentUserHandler)
			protected.Post("/regenerate-api-key", handlers.RegenerateAPIKeyHandler)

			protected.Post("/insert", handlers.InsertHandler)
			protected.Get("/getModels", handlers.ReadHandler)
			if deleteModelHandler != nil {
				protected.Delete("/deleteModel", deleteModelHandler.DeleteModel)
			}
			protected.Get("/downloadModel", handlers.DownloadTrainedModelHandler)

			// Community marketplace routes
			protected.Post("/publish", handlers.PubHandler)
			protected.Post("/published-models/{id}/unpublish", handlers.UnPublishModel)
			protected.Get("/published-models", handlers.GetPublishedModelsHandler)
			protected.Get("/my-published-models", handlers.GetMyPublishedModelsHandler)
			protected.Get("/published-models/{id}", handlers.GetPublishedModelByIDHandler)
			protected.Post("/published-models/{id}/download", handlers.DownloadPublishedModelHandler)

			// Likes
			protected.Post("/published-models/{id}/like", handlers.LikeModelHandler)
			protected.Delete("/published-models/{id}/like", handlers.UnlikeModelHandler)
			protected.Get("/published-models/{id}/likes", handlers.GetModelLikesHandler)

			// Comments
			protected.Get("/published-models/{id}/comments", handlers.GetModelCommentsHandler)
			protected.Post("/published-models/{id}/comments", handlers.AddModelCommentHandler)
			protected.Delete("/comments/{commentId}", handlers.DeleteModelCommentHandler)

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

			// Subscription routes
			protected.Get("/subscription", handlers.GetSubscriptionHandler)
			protected.Post("/subscription/checkout", handlers.CreateCheckoutSessionHandler)
			protected.Post("/subscription/mock-upgrade", handlers.MockUpgradeHandler) // For development/testing only
			protected.Get("/pricing", handlers.GetPricingHandler)

			// Agent status
			protected.Get("/agent/status", handlers.GetAgentStatusHandler)

			// HuggingFace integration routes
			protected.Post("/huggingface/push", handlers.PushToHuggingFaceHandler)
			protected.Post("/huggingface/import", handlers.ImportFromHuggingFaceHandler)
			protected.Post("/huggingface/inference", handlers.RunHuggingFaceInferenceHandler)
		})

		// Public HuggingFace search (no auth required, but token optional)
		r.Get("/huggingface/search", handlers.SearchHuggingFaceModelsHandler)
		r.Post("/huggingface/search", handlers.SearchHuggingFaceModelsHandler)

		// Public webhook endpoint (no auth required)
		r.Post("/webhook/stripe", handlers.StripeWebhookHandler)

		// Public pricing endpoint
		r.Get("/pricing", handlers.GetPricingHandler)
	})
	return r


}






