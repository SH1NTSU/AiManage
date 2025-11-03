package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/BradPerbs/claude-go"
	"github.com/joho/godotenv"
)

func main() {
	
	godotenv.Load()
	
	key := os.Getenv("ANTHROPIC_API_KEY");
	
	client := claude.NewClient(key)

	// Send a prompt to Claude
	prompt := "What is the capital of France?"
	response, err := client.SendPrompt(prompt)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(response)
}

