package ollama

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/1steve78/mcp-hub/internal/config"
)

func TestClient_Chat(t *testing.T) {
	tests := []struct {
		name           string
		timeout        time.Duration
		handler        http.HandlerFunc
		wantReply      string
		wantErrContain string
		wantErrIs      error
	}{
		{
			name:    "success path",
			timeout: 5 * time.Second,
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/chat" {
					t.Errorf("expected /api/chat, got %s", r.URL.Path)
				}
				if ct := r.Header.Get("Content-Type"); ct != "application/json" {
					t.Errorf("expected Content-Type application/json, got %q", ct)
				}

				var req ChatRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Fatalf("failed to decode request: %v", err)
				}
				if req.Stream {
					t.Errorf("expected stream=false, got true")
				}
				if req.Model != "qwen3:1.7b" {
					t.Errorf("expected model qwen3:1.7b, got %q", req.Model)
				}
				if len(req.Messages) != 2 || req.Messages[1].Content != "Hello" {
					t.Errorf("unexpected messages payload: %+v", req.Messages)
				}

				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(ChatResponse{
					Model:   req.Model,
					Message: Message{Role: RoleAssistant, Content: "Mocked reply"},
					Done:    true,
				})
			},
			wantReply: "Mocked reply",
		},
		{
			name:    "empty content returns error",
			timeout: 5 * time.Second,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(ChatResponse{
					Model:   "qwen3:1.7b",
					Message: Message{Role: RoleAssistant, Content: "   "},
					Done:    true,
				})
			},
			wantErrContain: "empty response",
		},
		{
			name:    "HTTP 404 with JSON error",
			timeout: 5 * time.Second,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(ChatResponse{
					Error: "model 'xyz' not found",
				})
			},
			wantErrContain: "model 'xyz' not found",
		},
		{
			name:    "HTTP 500 with plain text body",
			timeout: 5 * time.Second,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("internal server failure"))
			},
			wantErrContain: "500",
		},
		{
			name:    "malformed JSON returns decode error",
			timeout: 5 * time.Second,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{invalid-json`))
			},
			wantErrContain: "decode",
		},
		{
			name:    "timeout returns context.DeadlineExceeded",
			timeout: 30 * time.Millisecond,
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(150 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
			},
			wantErrIs: context.DeadlineExceeded,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(tc.handler)
			defer srv.Close()

			client := NewClient(config.OllamaConfig{
				BaseURL:        srv.URL,
				DefaultModel:   "qwen3:1.7b",
				TimeoutSeconds: 5,
			})

			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()

			messages := []Message{
				{Role: RoleSystem, Content: "You are MCP Hub."},
				{Role: RoleUser, Content: "Hello"},
			}

			got, err := client.Chat(ctx, "qwen3:1.7b", messages)

			if tc.wantErrContain != "" || tc.wantErrIs != nil {
				if err == nil {
					t.Fatalf("expected error, got nil (reply=%q)", got)
				}
				if tc.wantErrContain != "" && !strings.Contains(err.Error(), tc.wantErrContain) {
					t.Errorf("expected error containing %q, got %v", tc.wantErrContain, err)
				}
				if tc.wantErrIs != nil && !errors.Is(err, tc.wantErrIs) {
					t.Errorf("expected errors.Is(%v), got %v", tc.wantErrIs, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantReply {
				t.Errorf("expected reply %q, got %q", tc.wantReply, got)
			}
		})
	}
}
