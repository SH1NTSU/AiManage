package aiAgent

import "time"

// DirectoryInfo represents information about a directory
type DirectoryInfo struct {
	Name         string       `json:"name"`
	Path         string       `json:"path"`
	Files        []FileInfo   `json:"files"`
	Subdirs      []string     `json:"subdirs"`
	TotalFiles   int          `json:"total_files"`
	TotalSize    int64        `json:"total_size"`
	LastModified time.Time    `json:"last_modified"`
}

// FileInfo represents information about a file
type FileInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	Extension string    `json:"extension"`
	Modified  time.Time `json:"modified"`
}

// AgentRequest represents a request to the AI agent
type AgentRequest struct {
	FolderName string `json:"folder_name"`
	Action     string `json:"action"` // "analyze", "train", "test"
	UserID     string `json:"user_id"`
}

// AgentResponse represents the AI agent's response
type AgentResponse struct {
	Success      bool                   `json:"success"`
	Message      string                 `json:"message"`
	DirectoryInfo *DirectoryInfo        `json:"directory_info,omitempty"`
	Statistics   map[string]interface{} `json:"statistics,omitempty"`
	Error        string                 `json:"error,omitempty"`
}
