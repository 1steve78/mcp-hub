package cli

import (
	"fmt"
	"strings"

	"github.com/1steve78/mcp-hub/internal/config"
	"github.com/1steve78/mcp-hub/internal/ollama"
)

func Run(args []string) error {

	out, err := process(args)

	if err != nil {
		if out != "" {
			fmt.Println(out)
		}
		return err
	}

	fmt.Println(out)

	return nil

}

func process(arg []string) (string, error) {

	if len(arg) == 0 {
		return printHelp(), nil
	}

	cfg, err := config.Load("config.yaml")
	if err != nil {
		return "", fmt.Errorf("config error: %w", err)
	}

	switch arg[0] {

	case "version", "-v", "--version":
		return fmt.Sprintf("%s v%s", cfg.App.Name, cfg.App.Version), nil

	case "models":
		return getOllamaModels(cfg.Ollama)

	case "chat":
		return "chat is called", nil

	case "help":
		return printHelp(), nil

	default:
		return printHelp(), fmt.Errorf("unknown command: %s", arg[0])
	}

}

func printHelp() string {
	return `MCP Hub - Local AI Developer Assistant

Usage: mcp-hub <command>

Commands:
  version   Print the current version
  models    List available models
  chat      Start a chat session
  help      Show this help message`
}

func getOllamaModels(cfg config.OllamaConfig) (string, error) {

	client := ollama.NewClient(cfg)

	models, err := client.ListModels()
	if err != nil {
		return "", fmt.Errorf("failed to list ollama models: %w", err)
	}

	if len(models) == 0 {
		return "No models found.", nil
	}

	result := ""
	for _, m := range models {
		result += fmt.Sprintf("- %s (Size: %d bytes)\n", m.Name, m.Size)
	}

	return strings.TrimRight(result, "\n"), nil
}
