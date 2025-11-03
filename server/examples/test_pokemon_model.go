package main

import (
	"fmt"
	"os"
	"server/aiAgent"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	godotenv.Load()

	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		fmt.Println("Note: GEMINI_API_KEY not set - Gemini AI analysis will be skipped")
	}

	// Get the uploads path
	uploadsPath := "./uploads"

	// Create a new AI agent
	agent, err := aiAgent.NewAgent(key, uploadsPath)
	if err != nil {
		fmt.Println("Error creating agent:", err)
		return
	}

	fmt.Println("🤖 AI Agent initialized successfully!")
	fmt.Println("================================================\n")

	// Test with the real PokemonModel directory
	pokemonDir := "PokemonModel"

	// Example 1: Get directory info
	fmt.Printf("📁 Getting info for directory: %s\n", pokemonDir)
	fmt.Println("================================================")
	infoReq := aiAgent.AgentRequest{
		FolderName: pokemonDir,
		Action:     "info",
	}
	infoResp, err := agent.ProcessRequest(infoReq)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("✅ Success: %v\n", infoResp.Success)
	fmt.Printf("📝 Message: %s\n", infoResp.Message)

	if infoResp.DirectoryInfo != nil {
		dirInfo := infoResp.DirectoryInfo
		fmt.Printf("\n📊 Directory Statistics:\n")
		fmt.Printf("   - Total files: %d\n", dirInfo.TotalFiles)
		fmt.Printf("   - Total size: %.2f MB\n", float64(dirInfo.TotalSize)/(1024*1024))
		fmt.Printf("   - Subdirectories: %d\n", len(dirInfo.Subdirs))

		// Show file type distribution
		fileTypes := make(map[string]int)
		for _, file := range dirInfo.Files {
			ext := file.Extension
			if ext == "" {
				ext = "no extension"
			}
			fileTypes[ext]++
		}

		fmt.Printf("\n📈 File Type Distribution:\n")
		for ext, count := range fileTypes {
			fmt.Printf("   - .%s: %d files\n", ext, count)
		}

		// Show some sample files
		if len(dirInfo.Files) > 0 {
			fmt.Printf("\n📄 Sample Files (first 5):\n")
			limit := 5
			if len(dirInfo.Files) < limit {
				limit = len(dirInfo.Files)
			}
			for i := 0; i < limit; i++ {
				file := dirInfo.Files[i]
				fmt.Printf("   - %s (%.2f KB)\n", file.Name, float64(file.Size)/1024)
			}
		}
	}

	fmt.Println("\n================================================")

	// Example 2: Analyze directory with Gemini AI (if API key is set)
	if key != "" {
		fmt.Printf("\n🧠 Analyzing directory with Gemini AI: %s\n", pokemonDir)
		fmt.Println("================================================")

		analyzeReq := aiAgent.AgentRequest{
			FolderName: pokemonDir,
			Action:     "analyze",
		}
		analyzeResp, err := agent.ProcessRequest(analyzeReq)
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Printf("✅ Success: %v\n\n", analyzeResp.Success)
			if analyzeResp.Message != "" {
				fmt.Printf("📋 AI Analysis:\n%s\n", analyzeResp.Message)
			}
			if analyzeResp.Error != "" {
				fmt.Printf("\n⚠️  Note: %s\n", analyzeResp.Error)
			}
		}
	} else {
		fmt.Println("\n⏭️  Skipping AI analysis (no API key)")
	}

	fmt.Println("\n================================================")
	fmt.Println("✅ Test completed successfully!")
}
