package llm

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/liliang-cn/ohman/internal/config"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/shared"
)

// StreamHandler is a callback function for handling streaming chunks
type StreamHandler func(chunk string)

// Client is the LLM client interface
type Client interface {
	Chat(messages []Message) (*Response, error)
	ChatStream(messages []Message, handler StreamHandler) (*Response, error)
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Response represents an LLM response
type Response struct {
	Content      string
	TokensUsed   int
	FinishReason string
}

// NewClient creates an OpenAI-compatible LLM client
// Supports OpenAI and any OpenAI-compatible API endpoints
func NewClient(cfg config.LLMConfig) (Client, error) {
	return NewOpenAIClient(cfg)
}

// OpenAIClient is the OpenAI-compatible API client
type OpenAIClient struct {
	cfg    config.LLMConfig
	client openai.Client
}

// NewOpenAIClient creates an OpenAI-compatible client
// Works with OpenAI, Azure OpenAI, and any OpenAI-compatible endpoints
func NewOpenAIClient(cfg config.LLMConfig) (*OpenAIClient, error) {
	opts := []option.RequestOption{
		option.WithAPIKey(cfg.APIKey),
	}

	// Set base URL if provided (for custom/compatible endpoints)
	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}

	if cfg.Timeout > 0 {
		opts = append(opts, option.WithHTTPClient(&http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		}))
	}

	client := openai.NewClient(opts...)

	return &OpenAIClient{
		cfg:    cfg,
		client: client,
	}, nil
}

// Chat sends a chat request with streaming enabled by default
func (c *OpenAIClient) Chat(messages []Message) (*Response, error) {
	// Use streaming with a handler that prints each chunk
	return c.ChatStream(messages, func(chunk string) {
		fmt.Print(chunk)
	})
}

// ChatStream sends a streaming chat request
func (c *OpenAIClient) ChatStream(messages []Message, handler StreamHandler) (*Response, error) {
	ctx := context.Background()

	// Convert messages to OpenAI format
	chatMessages := make([]openai.ChatCompletionMessageParamUnion, len(messages))
	for i, msg := range messages {
		switch msg.Role {
		case "system":
			chatMessages[i] = openai.SystemMessage(msg.Content)
		case "user":
			chatMessages[i] = openai.UserMessage(msg.Content)
		case "assistant":
			chatMessages[i] = openai.AssistantMessage(msg.Content)
		default:
			chatMessages[i] = openai.UserMessage(msg.Content)
		}
	}

	params := openai.ChatCompletionNewParams{
		Messages: chatMessages,
		Model:    shared.ChatModel(c.cfg.Model),
	}

	if c.cfg.MaxTokens > 0 {
		params.MaxTokens = param.NewOpt(int64(c.cfg.MaxTokens))
	}

	if c.cfg.Temperature > 0 {
		params.Temperature = param.NewOpt(c.cfg.Temperature)
	}

	stream := c.client.Chat.Completions.NewStreaming(ctx, params)

	var fullContent string
	var finishReason string

	for stream.Next() {
		chunk := stream.Current()
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			if delta != "" {
				fullContent += delta
				if handler != nil {
					handler(delta)
				}
			}
			if chunk.Choices[0].FinishReason != "" {
				finishReason = string(chunk.Choices[0].FinishReason)
			}
		}
	}

	if err := stream.Err(); err != nil {
		return nil, fmt.Errorf("streaming error: %w", err)
	}

	if fullContent == "" {
		return nil, fmt.Errorf("no response received")
	}

	return &Response{
		Content:      fullContent,
		TokensUsed:   0, // Streaming doesn't provide token count in real-time
		FinishReason: finishReason,
	}, nil
}
