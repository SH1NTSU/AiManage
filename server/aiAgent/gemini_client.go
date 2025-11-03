package aiAgent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	// Google AI Studio API endpoint
	geminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/models"
	defaultModel = "gemini-1.5-flash" // Widely available model
)

// GeminiClient handles communication with Google's Gemini API
type GeminiClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewGeminiClient creates a new Gemini API client
func NewGeminiClient(apiKey string) *GeminiClient {
	return &GeminiClient{
		apiKey:     apiKey,
		model:      defaultModel,
		httpClient: &http.Client{},
	}
}

// NewGeminiClientWithModel creates a new Gemini API client with a specific model
func NewGeminiClientWithModel(apiKey, model string) *GeminiClient {
	return &GeminiClient{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{},
	}
}

// GeminiRequest represents a request to the Gemini API
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

// GeminiContent represents content in the request
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

// GeminiPart represents a part of the content
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiResponse represents a response from the Gemini API
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		} `json:"content"`
		FinishReason  string `json:"finishReason"`
		Index         int    `json:"index"`
		SafetyRatings []struct {
			Category    string `json:"category"`
			Probability string `json:"probability"`
		} `json:"safetyRatings"`
	} `json:"candidates"`
	PromptFeedback struct {
		SafetyRatings []struct {
			Category    string `json:"category"`
			Probability string `json:"probability"`
		} `json:"safetyRatings"`
	} `json:"promptFeedback"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

// GeminiErrorResponse represents an error from the Gemini API
type GeminiErrorResponse struct {
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// SendPrompt sends a prompt to Gemini and returns the response
func (c *GeminiClient) SendPrompt(prompt string) (string, error) {
	// Create the request payload
	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: prompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build the URL with API key as query parameter
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", geminiAPIURL, c.model, c.apiKey)

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errResp GeminiErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return "", fmt.Errorf("API error (%d): %s - %s", errResp.Error.Code, errResp.Error.Status, errResp.Error.Message)
	}

	// Parse successful response
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract text from response
	if len(geminiResp.Candidates) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	if len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content in Gemini response")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// SendPromptWithHistory sends a prompt with conversation history
func (c *GeminiClient) SendPromptWithHistory(messages []GeminiContent) (string, error) {
	reqBody := GeminiRequest{
		Contents: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:generateContent?key=%s", geminiAPIURL, c.model, c.apiKey)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp GeminiErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		return "", fmt.Errorf("API error (%d): %s - %s", errResp.Error.Code, errResp.Error.Status, errResp.Error.Message)
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}
