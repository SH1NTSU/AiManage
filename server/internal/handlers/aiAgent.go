package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"server/aiAgent"
)

// AIAgentHandler handles AI agent requests
type AIAgentHandler struct {
	agent *aiAgent.Agent
}

// GetAgent returns the underlying agent (for use by other handlers)
func (h *AIAgentHandler) GetAgent() *aiAgent.Agent {
	return h.agent
}

// NewAIAgentHandler creates a new AI agent handler
func NewAIAgentHandler() (*AIAgentHandler, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, http.ErrAbortHandler
	}

	// Get the uploads path relative to the server root
	uploadsPath := filepath.Join(".", "uploads")

	agent, err := aiAgent.NewAgent(apiKey, uploadsPath)
	if err != nil {
		return nil, err
	}

	return &AIAgentHandler{
		agent: agent,
	}, nil
}

// AnalyzeDirectory handles directory analysis requests
func (h *AIAgentHandler) AnalyzeDirectory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req aiAgent.AgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.FolderName == "" {
		http.Error(w, "folder_name is required", http.StatusBadRequest)
		return
	}

	// Default action is analyze
	if req.Action == "" {
		req.Action = "analyze"
	}

	response, err := h.agent.ProcessRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetDirectoryInfo handles requests to get directory information
func (h *AIAgentHandler) GetDirectoryInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	folderName := r.URL.Query().Get("folder")
	if folderName == "" {
		http.Error(w, "folder query parameter is required", http.StatusBadRequest)
		return
	}

	req := aiAgent.AgentRequest{
		FolderName: folderName,
		Action:     "info",
	}

	response, err := h.agent.ProcessRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListDirectories handles requests to list all directories
func (h *AIAgentHandler) ListDirectories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req := aiAgent.AgentRequest{
		Action: "list",
	}

	response, err := h.agent.ProcessRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CustomPrompt handles custom prompt requests
func (h *AIAgentHandler) CustomPrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestBody struct {
		FolderName string `json:"folder_name"`
		Prompt     string `json:"prompt"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if requestBody.FolderName == "" || requestBody.Prompt == "" {
		http.Error(w, "folder_name and prompt are required", http.StatusBadRequest)
		return
	}

	response, err := h.agent.AnalyzeWithPrompt(requestBody.FolderName, requestBody.Prompt)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": response,
	})
}
