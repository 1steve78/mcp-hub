package cli

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/1steve78/mcp-hub/internal/config"
	"github.com/1steve78/mcp-hub/internal/ollama"
)

func runChat(client *ollama.Client, ollamaCfg config.OllamaConfig, args []string) (string, error) {
	fs := flag.NewFlagSet("chat", flag.ContinueOnError)
	modelFlag := fs.String("model", ollamaCfg.DefaultModel, "ollama model to use")
	fs.StringVar(modelFlag, "m", ollamaCfg.DefaultModel, "Ollama model to use (shorthand)")

	if err := fs.Parse(args); err != nil {
		return "", err
	}

	model := strings.TrimSpace(*modelFlag)
	if model == "" {
		return "", errors.New("model name cannot be empty (set ollama.default_model in config.yaml or pass --model)")
	}

	timeout := time.Duration(ollamaCfg.TimeoutSeconds) * time.Second

	if prompt := strings.TrimSpace(strings.Join(fs.Args(), " ")); prompt != "" {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		messages := []ollama.Message{
			{Role: ollama.RoleUser, Content: prompt},
		}
		return client.Chat(ctx, model, messages)
	}

	return runChatREPL(client, model, timeout)

}

func runChatREPL(client *ollama.Client, model string, timeout time.Duration) (string, error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		_ = os.Stdin.Close()
	}()

	history := newChatHistory()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You > ")
		if !scanner.Scan() {
			if ctx.Err() != nil {
				fmt.Println("\nSession ended.")
				return "", nil
			}
			if err := scanner.Err(); err != nil {
				return "", fmt.Errorf("input error: %w", err)
			}

			fmt.Println("\nSession ended.")
			return "", nil
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		switch input {
		case "/exit", "/quit":
			fmt.Println("\nSession ended.")
			return "", nil

		case "/clear":
			history = newChatHistory()
			fmt.Print("Conversation cleared.\n\n")
			continue
		}

		candidateHistory := make([]ollama.Message, len(history)+1)
		copy(candidateHistory, history)
		candidateHistory[len(history)] = ollama.Message{
			Role:    ollama.RoleUser,
			Content: input,
		}

		reCtx, cancel := context.WithTimeout(ctx, timeout)
		start := time.Now()
		reply, err := client.Chat(reCtx, model, candidateHistory)
		elapsed := time.Since(start)
		cancel()

		if err != nil {
			if ctx.Err() != nil {
				fmt.Println("\nSession ended.")
				return "", nil
			}

			fmt.Printf("Error: %v\n\n", err)
			continue
		}

		history = append(candidateHistory, ollama.Message{
			Role:    ollama.RoleAssistant,
			Content: reply,
		})

		fmt.Printf("\nMCP Hub > %s\n", reply)
		fmt.Printf("Response received (%.2fs)\n\n", elapsed.Seconds())
	}
}

func newChatHistory() []ollama.Message {
	return []ollama.Message{
		{
			Role:    ollama.RoleSystem,
			Content: "You are MCP Hub, a helpful local-first coding assistant.",
		},
	}
}
