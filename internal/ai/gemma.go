package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	defaultOllamaBaseURL = "http://127.0.0.1:11434"
	modelName            = "gemma:2b"
	auditPrompt          = "Act as a Senior DevSecOps Security Auditor. Analyze the following Terraform configuration and provide 3 key Zero-Trust recommendations or security improvements:\n\n"
	fallbackNotice       = "ℹ️ No active AI engine detected (Ollama or API Key). Security audit skipped."
)

// Configurable endpoints for testing
var (
	ollamaBaseURL  = defaultOllamaBaseURL
	geminiEndpoint = "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent"
	openAIEndpoint = "https://api.openai.com/v1/chat/completions"
)

type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type generateResponse struct {
	Model    string `json:"model"`
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type openAIRequest struct {
	Model    string          `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// CheckOllamaAvailable performs a GET to http://127.0.0.1:11434/ with a short 3-second timeout.
// If it responds with HTTP 200 OK, it assumes Ollama is active.
func CheckOllamaAvailable() bool {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	resp, err := client.Get(ollamaBaseURL + "/")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// Audit executes the Terraform security audit with automatic detection:
// 1. Checks Ollama availability at http://127.0.0.1:11434/ (timeout 3s).
// 2. If active, sends POST to http://127.0.0.1:11434/api/generate with gemma:2b (timeout 60s).
//    If the call fails, prints error in console for debugging.
// 3. If Ollama is unavailable or fails, checks GEMINI_API_KEY or OPENAI_API_KEY.
// 4. If no AI engine responds, applies graceful degradation returning an informational notice.
func Audit(tfContent string) (string, error) {
	// 1. Ollama availability check
	if CheckOllamaAvailable() {
		resp, err := auditOllama(tfContent)
		if err == nil && resp != "" {
			return resp, nil
		}
		// Print exact error to console if the call fails for debugging
		fmt.Printf("⚠️  [Ollama] Error calling model '%s': %v\n", modelName, err)
	}

	// 2. Fallback to Gemini if GEMINI_API_KEY is configured
	if geminiKey := os.Getenv("GEMINI_API_KEY"); geminiKey != "" {
		if resp, err := auditGemini(geminiKey, tfContent); err == nil && resp != "" {
			return resp, nil
		} else if err != nil {
			fmt.Printf("⚠️  [Gemini] Error calling API: %v\n", err)
		}
	}

	// 3. Fallback to OpenAI if OPENAI_API_KEY is configured
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
		if resp, err := auditOpenAI(openaiKey, tfContent); err == nil && resp != "" {
			return resp, nil
		} else if err != nil {
			fmt.Printf("⚠️  [OpenAI] Error calling API: %v\n", err)
		}
	}

	// 4. Graceful degradation (Safe Fallback)
	return fallbackNotice, nil
}

// AuditTerraform is an alias of Audit for backwards compatibility
func AuditTerraform(tfContent string) (string, error) {
	return Audit(tfContent)
}

func auditOllama(tfContent string) (string, error) {
	reqBody := generateRequest{
		Model:  modelName,
		Prompt: auditPrompt + tfContent,
		Stream: false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error serializing request for Ollama: %w", err)
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	generateURL := ollamaBaseURL + "/api/generate"
	resp, err := client.Post(generateURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("connection error with Ollama at %s: %w", generateURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama returned HTTP status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var genResp generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return "", fmt.Errorf("error decoding Ollama response: %w", err)
	}

	if genResp.Error != "" {
		return "", fmt.Errorf("error reported by Ollama: %s", genResp.Error)
	}

	return genResp.Response, nil
}

func auditGemini(apiKey, tfContent string) (string, error) {
	url := fmt.Sprintf("%s?key=%s", geminiEndpoint, apiKey)

	reqData := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: auditPrompt + tfContent},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return "", err
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Gemini API HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var geminiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", err
	}

	if geminiResp.Error != nil && geminiResp.Error.Message != "" {
		return "", fmt.Errorf("Gemini error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return geminiResp.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("empty response from Gemini API")
}

func auditOpenAI(apiKey, tfContent string) (string, error) {
	url := openAIEndpoint

	reqData := openAIRequest{
		Model: "gpt-4o-mini",
		Messages: []openAIMessage{
			{Role: "user", Content: auditPrompt + tfContent},
		},
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var oaiResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&oaiResp); err != nil {
		return "", err
	}

	if oaiResp.Error != nil && oaiResp.Error.Message != "" {
		return "", fmt.Errorf("OpenAI error: %s", oaiResp.Error.Message)
	}

	if len(oaiResp.Choices) > 0 {
		return oaiResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("empty response from OpenAI API")
}
