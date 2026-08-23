package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/alifatolahi/ganjoor-mcp/internal/bridge"
)

func main() {
	model := flag.String("model", "qwen3", "Ollama model name")
	ollamaURL := flag.String("ollama-url", "http://localhost:11434", "Ollama API URL")
	mcpCommand := flag.String("mcp-command", "go", "MCP server executable")
	mcpRun := flag.Bool("mcp-run", true, "run the MCP command as `go run ./cmd/ganjoor-mcp`")
	flag.Parse()

	args := defaultMCPArgs()
	if !*mcpRun {
		args = flag.Args()
	}
	if err := (bridge.Bridge{Config: bridge.Config{
		OllamaURL: *ollamaURL, Model: *model, MCPCommand: *mcpCommand, MCPArguments: args,
	}}).Run(context.Background(), os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

func defaultMCPArgs() []string {
	return []string{"run", "./cmd/ganjoor-mcp"}
}
