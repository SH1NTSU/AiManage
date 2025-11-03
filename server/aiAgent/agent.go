package aiAgent

import (
	"fmt"
	"os"
	"strings"
)

// Agent represents the AI agent with Gemini integration
type Agent struct {
	client    *GeminiClient
	navigator *DirectoryNavigator
	trainer   *Trainer
	apiKey    string
}

// NewAgent creates a new AI agent instance
func NewAgent(apiKey string, uploadsPath string) (*Agent, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required")
	}

	// Ensure uploads directory exists
	if err := os.MkdirAll(uploadsPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create uploads directory: %w", err)
	}

	client := NewGeminiClient(apiKey)
	navigator := NewDirectoryNavigator(uploadsPath)
	trainer := NewTrainer(navigator)

	return &Agent{
		client:    client,
		navigator: navigator,
		trainer:   trainer,
		apiKey:    apiKey,
	}, nil
}

// ProcessRequest processes an agent request
func (a *Agent) ProcessRequest(req AgentRequest) (*AgentResponse, error) {
	switch req.Action {
	case "analyze":
		return a.analyzeDirectory(req.FolderName)
	case "list":
		return a.listDirectories()
	case "info":
		return a.getDirectoryInfo(req.FolderName)
	default:
		return &AgentResponse{
			Success: false,
			Error:   fmt.Sprintf("unknown action: %s", req.Action),
		}, nil
	}
}

// analyzeDirectory analyzes a directory using Claude AI
func (a *Agent) analyzeDirectory(folderName string) (*AgentResponse, error) {
	// First, get directory info
	dirInfo, err := a.navigator.OpenDirectory(folderName)
	if err != nil {
		return &AgentResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Prepare a summary for Claude
	summary := a.prepareDirectorySummary(dirInfo)

	// Send to Claude for analysis
	prompt := fmt.Sprintf(`Analyze the following directory structure and provide insights:

%s

Please provide:
1. A brief overview of the directory contents
2. File type distribution
3. Any patterns you notice
4. Suggestions for organization or potential use cases
5. If this looks like a dataset, what kind of machine learning task it might be suitable for

Keep your response concise and actionable.`, summary)

	response, err := a.client.SendPrompt(prompt)
	if err != nil {
		return &AgentResponse{
			Success:       true, // We still got directory info
			DirectoryInfo: dirInfo,
			Message:       "Directory info retrieved, but AI analysis failed",
			Error:         err.Error(),
		}, nil
	}

	return &AgentResponse{
		Success:       true,
		Message:       response,
		DirectoryInfo: dirInfo,
	}, nil
}

// getDirectoryInfo returns directory information without AI analysis
func (a *Agent) getDirectoryInfo(folderName string) (*AgentResponse, error) {
	dirInfo, err := a.navigator.OpenDirectory(folderName)
	if err != nil {
		return &AgentResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &AgentResponse{
		Success:       true,
		Message:       fmt.Sprintf("Successfully read directory '%s'", folderName),
		DirectoryInfo: dirInfo,
	}, nil
}

// listDirectories lists all available directories
func (a *Agent) listDirectories() (*AgentResponse, error) {
	dirs, err := a.navigator.ListDirectories()
	if err != nil {
		return &AgentResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	statistics := map[string]interface{}{
		"available_directories": dirs,
		"total_count":          len(dirs),
	}

	return &AgentResponse{
		Success:    true,
		Message:    fmt.Sprintf("Found %d directories", len(dirs)),
		Statistics: statistics,
	}, nil
}

// prepareDirectorySummary prepares a text summary of directory contents
func (a *Agent) prepareDirectorySummary(dirInfo *DirectoryInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Directory: %s\n", dirInfo.Name))
	sb.WriteString(fmt.Sprintf("Path: %s\n", dirInfo.Path))
	sb.WriteString(fmt.Sprintf("Total Files: %d\n", dirInfo.TotalFiles))
	sb.WriteString(fmt.Sprintf("Total Size: %.2f MB\n", float64(dirInfo.TotalSize)/(1024*1024)))
	sb.WriteString(fmt.Sprintf("Subdirectories: %d\n\n", len(dirInfo.Subdirs)))

	// File type distribution
	fileTypes := make(map[string]int)
	for _, file := range dirInfo.Files {
		ext := file.Extension
		if ext == "" {
			ext = "no extension"
		}
		fileTypes[ext]++
	}

	sb.WriteString("File Type Distribution:\n")
	for ext, count := range fileTypes {
		sb.WriteString(fmt.Sprintf("  - %s: %d files\n", ext, count))
	}

	// List subdirectories
	if len(dirInfo.Subdirs) > 0 {
		sb.WriteString("\nSubdirectories:\n")
		for _, subdir := range dirInfo.Subdirs {
			sb.WriteString(fmt.Sprintf("  - %s\n", subdir))
		}
	}

	// Sample files (first 10)
	if len(dirInfo.Files) > 0 {
		sb.WriteString("\nSample Files:\n")
		limit := 10
		if len(dirInfo.Files) < limit {
			limit = len(dirInfo.Files)
		}
		for i := 0; i < limit; i++ {
			file := dirInfo.Files[i]
			sb.WriteString(fmt.Sprintf("  - %s (%.2f KB)\n", file.Name, float64(file.Size)/1024))
		}
		if len(dirInfo.Files) > limit {
			sb.WriteString(fmt.Sprintf("  ... and %d more files\n", len(dirInfo.Files)-limit))
		}
	}

	return sb.String()
}

// OpenDirectory is a convenience method to directly open a directory
func (a *Agent) OpenDirectory(folderName string) (*DirectoryInfo, error) {
	return a.navigator.OpenDirectory(folderName)
}

// GetNavigator returns the directory navigator
func (a *Agent) GetNavigator() *DirectoryNavigator {
	return a.navigator
}

// GetTrainer returns the trainer
func (a *Agent) GetTrainer() *Trainer {
	return a.trainer
}

// AnalyzeWithPrompt sends a custom prompt to Claude about a directory
func (a *Agent) AnalyzeWithPrompt(folderName, customPrompt string) (string, error) {
	dirInfo, err := a.navigator.OpenDirectory(folderName)
	if err != nil {
		return "", err
	}

	summary := a.prepareDirectorySummary(dirInfo)
	fullPrompt := fmt.Sprintf("%s\n\nDirectory Information:\n%s", customPrompt, summary)

	response, err := a.client.SendPrompt(fullPrompt)
	if err != nil {
		return "", fmt.Errorf("gemini API error: %w", err)
	}

	return response, nil
}
