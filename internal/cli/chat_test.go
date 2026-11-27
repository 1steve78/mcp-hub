package cli

import (
	"testing"

	"github.com/1steve78/mcp-hub/internal/ollama"
)

func TestNewChatHistory_ClearReset(t *testing.T) {
	history := newChatHistory()
	history = append(history,
		ollama.Message{Role: ollama.RoleUser, Content: "Hi"},
		ollama.Message{Role: ollama.RoleAssistant, Content: "Hello!"},
	)

	if len(history) != 3 {
		t.Fatalf("expected 3 messages before clear, got %d", len(history))
	}

	// Simulate /clear reset
	history = newChatHistory()

	if len(history) != 1 {
		t.Fatalf("expected len(history) == 1 after clear, got %d", len(history))
	}
	if history[0].Role != ollama.RoleSystem {
		t.Errorf("expected history[0].Role == %q, got %q", ollama.RoleSystem, history[0].Role)
	}
}
