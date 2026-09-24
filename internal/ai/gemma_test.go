package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestCheckOllamaAvailable_True(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("Ollama is running"))
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	origBase := ollamaBaseURL
	ollamaBaseURL = server.URL
	defer func() { ollamaBaseURL = origBase }()

	if !CheckOllamaAvailable() {
		t.Errorf("Expected CheckOllamaAvailable to return true")
	}
}

func TestCheckOllamaAvailable_False(t *testing.T) {
	origBase := ollamaBaseURL
	ollamaBaseURL = "http://127.0.0.1:54321" // Non-existent port
	defer func() { ollamaBaseURL = origBase }()

	if CheckOllamaAvailable() {
		t.Errorf("Expected CheckOllamaAvailable to return false")
	}
}

func TestAudit_OllamaSuccess(t *testing.T) {
	expectedResponse := "1. Enable VPC Service Controls.\n2. Restrict IPs.\n3. Configure mTLS."

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.URL.Path == "/api/generate" {
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST method, got %s", r.Method)
			}

			var req generateRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Error decoding body: %v", err)
			}

			if req.Model != "gemma:2b" {
				t.Errorf("Expected model 'gemma:2b', got '%s'", req.Model)
			}
			if req.Stream != false {
				t.Errorf("Expected stream false, got %v", req.Stream)
			}
			if !strings.Contains(req.Prompt, "Act as a Senior DevSecOps Security Auditor") {
				t.Errorf("Prompt does not contain expected DevSecOps text")
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(generateResponse{
				Model:    "gemma:2b",
				Response: expectedResponse,
				Done:     true,
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	origBase := ollamaBaseURL
	ollamaBaseURL = server.URL
	defer func() { ollamaBaseURL = origBase }()

	result, err := Audit("resource google_cloud_run_v2_service api {}")
	if err != nil {
		t.Fatalf("Audit failed unexpectedly: %v", err)
	}

	if result != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, result)
	}
}

func TestAudit_GracefulDegradation(t *testing.T) {
	// Ollama unavailable
	origBase := ollamaBaseURL
	ollamaBaseURL = "http://127.0.0.1:54321"
	defer func() { ollamaBaseURL = origBase }()

	// Ensure no environment variables are set
	origGemini := os.Getenv("GEMINI_API_KEY")
	origOpenAI := os.Getenv("OPENAI_API_KEY")
	os.Unsetenv("GEMINI_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")
	defer func() {
		if origGemini != "" {
			os.Setenv("GEMINI_API_KEY", origGemini)
		}
		if origOpenAI != "" {
			os.Setenv("OPENAI_API_KEY", origOpenAI)
		}
	}()

	result, err := Audit("resource google_cloud_run_v2_service api {}")
	if err != nil {
		t.Fatalf("Audit should not return error on graceful fallback, got: %v", err)
	}

	expectedNotice := "ℹ️ No active AI engine detected (Ollama or API Key). Security audit skipped."
	if result != expectedNotice {
		t.Errorf("Expected response '%s', got '%s'", expectedNotice, result)
	}
}

func TestAudit_FallbackGemini(t *testing.T) {
	// Ollama down
	origBase := ollamaBaseURL
	ollamaBaseURL = "http://127.0.0.1:54321"
	defer func() { ollamaBaseURL = origBase }()

	expectedResponse := "Gemini: Use conditional IAM and Cloud Armor."

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := geminiResponse{
			Candidates: []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			}{
				{
					Content: struct {
						Parts []struct {
							Text string `json:"text"`
						} `json:"parts"`
					}{
						Parts: []struct {
							Text string `json:"text"`
						}{
							{Text: expectedResponse},
						},
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	origEndpoint := geminiEndpoint
	geminiEndpoint = server.URL
	defer func() { geminiEndpoint = origEndpoint }()

	os.Setenv("GEMINI_API_KEY", "test-key")
	defer os.Unsetenv("GEMINI_API_KEY")

	result, err := Audit("resource google_cloud_run_v2_service api {}")
	if err != nil {
		t.Fatalf("Audit failed unexpectedly: %v", err)
	}

	if !strings.Contains(result, "Cloud Armor") {
		t.Errorf("Expected Gemini response, got: %s", result)
	}
}

func TestAudit_FallbackOpenAI(t *testing.T) {
	// Ollama down
	origBase := ollamaBaseURL
	ollamaBaseURL = "http://127.0.0.1:54321"
	defer func() { ollamaBaseURL = origBase }()

	// Without Gemini
	os.Unsetenv("GEMINI_API_KEY")

	expectedResponse := "OpenAI: Principle of least privilege on Service Accounts."

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-openai-key" {
			t.Errorf("Incorrect Authorization header: %s", auth)
		}
		w.Header().Set("Content-Type", "application/json")
		resp := openAIResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{
					Message: struct {
						Content string `json:"content"`
					}{
						Content: expectedResponse,
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	origEndpoint := openAIEndpoint
	openAIEndpoint = server.URL
	defer func() { openAIEndpoint = origEndpoint }()

	os.Setenv("OPENAI_API_KEY", "test-openai-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	result, err := Audit("resource google_cloud_run_v2_service api {}")
	if err != nil {
		t.Fatalf("Audit failed unexpectedly: %v", err)
	}

	if !strings.Contains(result, "least privilege") {
		t.Errorf("Expected OpenAI response, got: %s", result)
	}
}
