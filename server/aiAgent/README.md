# AI Agent - Directory Navigation & Analysis

A Go-based AI agent that integrates with Claude AI to navigate, analyze, and interact with directories in the uploads folder.

## Features

- 📁 **Directory Navigation**: Open and explore specific folders by name
- 📊 **File Analysis**: Automatically scan and categorize files by type
- 🤖 **Claude AI Integration**: Get intelligent insights about your datasets
- 🔒 **Security**: Built-in path validation to prevent directory traversal attacks
- 📈 **Statistics**: File counts, size distribution, and structure analysis

## Project Structure

```
aiAgent/
├── agent.go          # Main AI agent with Claude integration
├── directory.go      # Directory navigation and file operations
├── types.go          # Data structures and type definitions
└── README.md         # This file
```

## API Endpoints

The AI agent is integrated into your server with these protected endpoints:

### 1. List All Directories
```bash
GET /v1/ai/directories
Authorization: Bearer <token>
```

**Response:**
```json
{
  "success": true,
  "message": "Found 2 directories",
  "statistics": {
    "available_directories": ["PokemonModel", "test_folder"],
    "total_count": 2
  }
}
```

### 2. Get Directory Info
```bash
GET /v1/ai/directory?folder=PokemonModel
Authorization: Bearer <token>
```

**Response:**
```json
{
  "success": true,
  "message": "Successfully read directory 'PokemonModel'",
  "directory_info": {
    "name": "PokemonModel",
    "path": "./uploads/PokemonModel",
    "total_files": 311,
    "total_size": 296828416,
    "subdirs": ["models", "data", "utils", ...],
    "files": [...]
  }
}
```

### 3. Analyze Directory with AI
```bash
POST /v1/ai/analyze
Authorization: Bearer <token>
Content-Type: application/json

{
  "folder_name": "PokemonModel",
  "action": "analyze"
}
```

**Response:**
```json
{
  "success": true,
  "message": "This appears to be a PyTorch machine learning project for Pokemon image classification. The directory contains training scripts, model definitions, dataset loaders, and a collection of Pokemon card images...",
  "directory_info": {...}
}
```

### 4. Custom AI Prompt
```bash
POST /v1/ai/prompt
Authorization: Bearer <token>
Content-Type: application/json

{
  "folder_name": "PokemonModel",
  "prompt": "What machine learning framework is this project using and what would be good metrics to track?"
}
```

## Setup

### 1. Environment Variables

Add your Anthropic API key to `.env`:
```bash
ANTHROPIC_API_KEY=your_api_key_here
```

### 2. Test the Agent

Run the standalone test:
```bash
go build -o test-pokemon ./examples/test_pokemon_model.go
./test-pokemon
```

## Usage in Your Application

### Initialize the Agent

```go
import "server/aiAgent"

agent, err := aiAgent.NewAgent(apiKey, "./uploads")
if err != nil {
    log.Fatal(err)
}
```

### List Directories

```go
req := aiAgent.AgentRequest{Action: "list"}
response, err := agent.ProcessRequest(req)
```

### Get Directory Info

```go
req := aiAgent.AgentRequest{
    FolderName: "PokemonModel",
    Action:     "info",
}
response, err := agent.ProcessRequest(req)
```

### Analyze with AI

```go
req := aiAgent.AgentRequest{
    FolderName: "PokemonModel",
    Action:     "analyze",
}
response, err := agent.ProcessRequest(req)
```

### Custom Analysis

```go
response, err := agent.AnalyzeWithPrompt(
    "PokemonModel",
    "Explain the training pipeline and suggest improvements",
)
```

## Security Features

- ✅ Path validation prevents directory traversal attacks
- ✅ All paths are confined to the uploads directory
- ✅ JWT authentication required for all API endpoints
- ✅ Read-only operations (no file modifications)

## Training Features ✅ IMPLEMENTED

The agent now includes full training capabilities:

### What's Included:

1. **✅ Run Training Scripts** - Execute Python training files
2. **✅ Monitor Progress** - Track training metrics in real-time
3. **✅ Generate Statistics** - Analyze model performance
4. **✅ Claude AI Analysis** - Get intelligent insights (optional)

### Training Components:

- **`trainer.go`** - Training execution and progress monitoring
- **`analyzer.go`** - Performance analysis (with/without AI)
- **`internal/handlers/training.go`** - HTTP API endpoints

### Quick Example:

```go
// Start training
trainer := agent.GetTrainer()
progress, err := trainer.StartTraining(ctx, aiAgent.TrainingRequest{
    FolderName:    "PokemonModel",
    ScriptName:    "train.py",
    PythonCommand: "python3",
})

// Get progress
progress, err = trainer.GetProgress(trainingID)

// Analyze results (no AI needed)
analysis := agent.QuickAnalysis(progress)

// Or with Claude AI insights
aiAnalysis, err := agent.AnalyzeTrainingResults(progress)
```

### See Full Guide:
📖 Check **[TRAINING_GUIDE.md](../TRAINING_GUIDE.md)** for complete API documentation, examples, and best practices.

## Tested With

- ✅ PokemonModel (PyTorch image classification project)
- ✅ 311 files, 283 MB
- ✅ Multiple subdirectories and file types

## Dependencies

```go
github.com/BradPerbs/claude-go  // Claude AI integration
github.com/joho/godotenv        // Environment variable loading
```

## Contributing

When adding new features:
1. Add new methods to `agent.go` for AI operations
2. Add new methods to `directory.go` for file operations
3. Update `types.go` for new data structures
4. Add new handlers in `internal/handlers/aiAgent.go`
5. Register routes in `internal/service/router.go`
