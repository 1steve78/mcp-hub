package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatResponse struct {
	Model     string  `json:"model"`
	CreatedAt string  `json:"created_at"`
	Message   Message `json:"message"`
	Done      bool    `json:"done"`
	Error     string  `json:"error,omitempty"`
}

func (c *Client) Chat(ctx context.Context, model string, message []Message) (string, error) {
	reqBody := ChatRequest{
		Model:    model,
		Messages: message,
		Stream:   false,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/chat", bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create chat request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("make sure Ollama is running (ollama serve): %w", err)

	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

		var errResp ChatResponse
		if json.Unmarshal(bodyBytes, &errResp) == nil && strings.TrimSpace(errResp.Error) != "" {
			return "", fmt.Errorf("ollama error (status %d): %s", resp.StatusCode, errResp.Error)
		}

		snippet := strings.TrimSpace(string(bodyBytes))
		if snippet != "" {
			return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, snippet)
		}

		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("failed to decode chat response: %w", err)
	}

	if strings.TrimSpace(chatResp.Message.Content) == "" {
		return "", errors.New("ollama returned an empty response")
	}

	return chatResp.Message.Content, nil
}
