package ollama

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/1steve78/mcp-hub/internal/config"
)

type Model struct {
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modified_at"`
}
type tagsResponse struct {
	Models []Model `json:"models"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(cfg config.OllamaConfig) *Client {

	return &Client{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.TimeoutSeconds) * time.Second,
		},
	}
}

func (c *Client) ListModels() ([]Model, error) {

	resp, err := c.httpClient.Get(c.baseURL + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("make sure Ollama is running (ollama serve): %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status : %d", resp.StatusCode)
	}

	var result tagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Models, nil
}
