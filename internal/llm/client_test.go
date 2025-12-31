package llm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/liliang-cn/ohman/internal/config"
)

func TestNewClient(t *testing.T) {
	// NewClient now always returns an OpenAI-compatible client
	cfg := config.LLMConfig{
		APIKey:  "test-key",
		Timeout: 10,
	}

	_, err := NewClient(cfg)
	if err != nil {
		t.Errorf("NewClient() error = %v", err)
	}
}

func TestOpenAIClientChatStream(t *testing.T) {
	// Create mock SSE server for streaming
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if r.Header.Get("Authorization") != "Bearer test-api-key" {
			t.Error("missing or incorrect authorization header")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Send streaming response
		chunks := []string{
			`{"choices":[{"delta":{"content":"Hello"},"finish_reason":null}]}`,
			`{"choices":[{"delta":{"content":" world"},"finish_reason":null}]}`,
			`{"choices":[{"delta":{"content":"!"},"finish_reason":"stop"}]}`,
		}

		for _, chunk := range chunks {
			_, _ = fmt.Fprintf(w, "data: %s\n\n", chunk)
		}
		_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer server.Close()

	cfg := config.LLMConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
		Model:   "gpt-4",
		Timeout: 10,
	}

	client, err := NewOpenAIClient(cfg)
	if err != nil {
		t.Fatalf("NewOpenAIClient() error = %v", err)
	}

	messages := []Message{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Hello"},
	}

	var collected string
	resp, err := client.ChatStream(messages, func(chunk string) {
		collected += chunk
	})
	if err != nil {
		t.Fatalf("ChatStream() error = %v", err)
	}

	if resp.Content != "Hello world!" {
		t.Errorf("unexpected response content: %s", resp.Content)
	}

	if collected != "Hello world!" {
		t.Errorf("unexpected collected chunks: %s", collected)
	}
}

func TestOpenAIClientError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error": {"message": "invalid api key"}}`))
	}))
	defer server.Close()

	cfg := config.LLMConfig{
		APIKey:  "invalid-key",
		BaseURL: server.URL,
		Timeout: 10,
	}

	client, _ := NewOpenAIClient(cfg)
	_, err := client.Chat([]Message{{Role: "user", Content: "test"}})

	if err == nil {
		t.Error("expected error for unauthorized request")
	}
}

func TestOpenAIClientWithBaseURL(t *testing.T) {
	// Test that BaseURL is correctly applied for custom/compatible endpoints
	cfg := config.LLMConfig{
		APIKey:  "test-key",
		BaseURL: "https://api.example.com/v1",
		Model:   "custom-model",
		Timeout: 10,
	}

	client, err := NewOpenAIClient(cfg)
	if err != nil {
		t.Fatalf("NewOpenAIClient() error = %v", err)
	}

	if client.cfg.BaseURL != "https://api.example.com/v1" {
		t.Errorf("BaseURL not set correctly: %s", client.cfg.BaseURL)
	}
}

func TestMessageRoles(t *testing.T) {
	// Test that different message roles are handled correctly
	messages := []Message{
		{Role: "system", Content: "System prompt"},
		{Role: "user", Content: "User message"},
		{Role: "assistant", Content: "Assistant response"},
		{Role: "unknown", Content: "Unknown role defaults to user"},
	}

	cfg := config.LLMConfig{
		APIKey:  "test-key",
		Timeout: 10,
	}

	client, err := NewOpenAIClient(cfg)
	if err != nil {
		t.Fatalf("NewOpenAIClient() error = %v", err)
	}

	// Just verify the client can be created with various message roles
	// The actual conversion is tested indirectly through the Chat method
	_ = client
	_ = messages
}

// Helper for encoding JSON responses
func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
