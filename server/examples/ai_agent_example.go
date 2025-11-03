package main

import (
	"fmt"
	"os"
	"path/filepath"
	"server/aiAgent"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	godotenv.Load()

	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		fmt.Println("Error: ANTHROPIC_API_KEY not set in environment")
		return
	}

	// Get the uploads path
	uploadsPath := filepath.Join("..", "uploads")

	// Create a new AI agent
	agent, err := aiAgent.NewAgent(key, uploadsPath)
	if err != nil {
		fmt.Println("Error creating agent:", err)
		return
	}

	fmt.Println("AI Agent initialized successfully!")
	fmt.Println("Available commands:")
	fmt.Println("  - List directories")
	fmt.Println("  - Analyze specific directory")
	fmt.Println("  - Get directory info")
	fmt.Println()

	// Example 1: List all directories
	fmt.Println("=== Listing all directories ===")
	listReq := aiAgent.AgentRequest{Action: "list"}
	listResp, err := agent.ProcessRequest(listReq)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Success: %v\n", listResp.Success)
	fmt.Printf("Message: %s\n", listResp.Message)
	if listResp.Statistics != nil {
		if dirs, ok := listResp.Statistics["available_directories"].([]string); ok {
			fmt.Println("Directories found:", dirs)
		}
	}
	fmt.Println()

	// Example 2: Create a test directory and file for demonstration
	testDir := "test_folder"
	fmt.Printf("=== Creating test directory: %s ===\n", testDir)
	navigator := agent.GetNavigator()
	if !navigator.DirectoryExists(testDir) {
		err = navigator.CreateDirectory(testDir)
		if err != nil {
			fmt.Println("Error creating test directory:", err)
		} else {
			fmt.Println("Test directory created successfully")
		}
	} else {
		fmt.Println("Test directory already exists")
	}
	fmt.Println()

	// Example 3: Get directory info
	fmt.Printf("=== Getting info for directory: %s ===\n", testDir)
	infoReq := aiAgent.AgentRequest{
		FolderName: testDir,
		Action:     "info",
	}
	infoResp, err := agent.ProcessRequest(infoReq)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Success: %v\n", infoResp.Success)
	fmt.Printf("Message: %s\n", infoResp.Message)
	if infoResp.DirectoryInfo != nil {
		fmt.Printf("Total files: %d\n", infoResp.DirectoryInfo.TotalFiles)
		fmt.Printf("Total size: %d bytes\n", infoResp.DirectoryInfo.TotalSize)
	}
	fmt.Println()

	// Example 4: Analyze directory (if files exist)
	fmt.Printf("=== Analyzing directory: %s ===\n", testDir)
	analyzeReq := aiAgent.AgentRequest{
		FolderName: testDir,
		Action:     "analyze",
	}
	analyzeResp, err := agent.ProcessRequest(analyzeReq)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Success: %v\n", analyzeResp.Success)
	if analyzeResp.Message != "" {
		fmt.Printf("AI Analysis:\n%s\n", analyzeResp.Message)
	}
	if analyzeResp.Error != "" {
		fmt.Printf("Note: %s\n", analyzeResp.Error)
	}
}

