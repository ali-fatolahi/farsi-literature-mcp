package bridge

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Config struct {
	OllamaURL    string
	Model        string
	MCPCommand   string
	MCPArguments []string
	MaxToolLoops int
}

type Bridge struct {
	Config Config
	Client *http.Client
	Logger *log.Logger
}

type ollamaMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content,omitempty"`
	ToolName  string           `json:"tool_name,omitempty"`
	ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
}

type ollamaToolCall struct {
	Function struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	} `json:"function"`
}

type ollamaTool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
		Parameters  any    `json:"parameters"`
	} `json:"function"`
}

type chatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Tools    []ollamaTool    `json:"tools,omitempty"`
	Stream   bool            `json:"stream"`
}

type chatResponse struct {
	Message ollamaMessage `json:"message"`
}

func (b Bridge) Run(ctx context.Context, input io.Reader, output io.Writer) error {
	cfg := b.Config
	if cfg.OllamaURL == "" {
		cfg.OllamaURL = "http://localhost:11434"
	}
	if cfg.Model == "" {
		cfg.Model = "qwen3"
	}
	if cfg.MCPCommand == "" {
		return errors.New("MCP command is required")
	}
	if cfg.MaxToolLoops < 1 {
		cfg.MaxToolLoops = 6
	}
	httpClient := b.Client
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	logger := b.Logger
	if logger == nil {
		logger = log.New(os.Stderr, "[ollama-mcp] ", log.LstdFlags)
	}

	logger.Printf("starting MCP server: %s %s", cfg.MCPCommand, strings.Join(cfg.MCPArguments, " "))
	cmd := exec.CommandContext(ctx, cfg.MCPCommand, cfg.MCPArguments...)
	cmd.Stderr = os.Stderr
	mcpIn, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("create MCP stdin: %w", err)
	}
	mcpOut, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create MCP stdout: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start MCP server: %w", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	mcpClient := mcp.NewClient(&mcp.Implementation{Name: "ollama-mcp-bridge", Version: "0.1.0"}, nil)
	session, err := mcpClient.Connect(ctx, &mcp.IOTransport{Reader: mcpOut, Writer: mcpIn}, nil)
	if err != nil {
		return fmt.Errorf("connect to MCP server: %w", err)
	}
	defer session.Close()
	logger.Printf("connected to MCP server")
	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		return fmt.Errorf("list MCP tools: %w", err)
	}

	ollamaTools, err := convertTools(tools.Tools)
	if err != nil {
		return err
	}
	logger.Printf("discovered %d MCP tools", len(ollamaTools))

	scanner := bufio.NewScanner(input)
	logger.Printf("ready for prompts (one per line)")
	for scanner.Scan() {
		prompt := strings.TrimSpace(scanner.Text())
		if prompt == "" {
			continue
		}
		logger.Printf("received prompt (%d characters)", len([]rune(prompt)))
		answer, err := b.chat(ctx, httpClient, session, cfg, ollamaTools, prompt, logger)
		if err != nil {
			return err
		}
		fmt.Fprintln(output, answer)
		logger.Printf("response complete")
	}
	return scanner.Err()
}

func (b Bridge) chat(ctx context.Context, client *http.Client, session *mcp.ClientSession, cfg Config, tools []ollamaTool, prompt string, logger *log.Logger) (string, error) {
	messages := []ollamaMessage{{Role: "user", Content: prompt}}
	for loop := 0; loop < cfg.MaxToolLoops; loop++ {
		logger.Printf("calling Ollama (round %d/%d)", loop+1, cfg.MaxToolLoops)
		response, err := b.callOllama(ctx, client, cfg, messages, tools)
		if err != nil {
			return "", err
		}
		messages = append(messages, response.Message)
		if len(response.Message.ToolCalls) == 0 {
			logger.Printf("Ollama returned final text")
			return response.Message.Content, nil
		}
		for _, call := range response.Message.ToolCalls {
			logger.Printf("calling MCP tool: %s", call.Function.Name)
			result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: call.Function.Name, Arguments: call.Function.Arguments})
			if err != nil {
				return "", fmt.Errorf("call MCP tool %q: %w", call.Function.Name, err)
			}
			content, err := json.Marshal(result)
			if err != nil {
				return "", fmt.Errorf("encode tool %q result: %w", call.Function.Name, err)
			}
			messages = append(messages, ollamaMessage{Role: "tool", ToolName: call.Function.Name, Content: string(content)})
		}
	}
	return "", errors.New("maximum Ollama tool-call loops exceeded")
}

func (b Bridge) callOllama(ctx context.Context, client *http.Client, cfg Config, messages []ollamaMessage, tools []ollamaTool) (chatResponse, error) {
	body, err := json.Marshal(chatRequest{Model: cfg.Model, Messages: messages, Tools: tools, Stream: false})
	if err != nil {
		return chatResponse{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(cfg.OllamaURL, "/")+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return chatResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return chatResponse{}, fmt.Errorf("call Ollama: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return chatResponse{}, fmt.Errorf("Ollama returned %s: %s", resp.Status, strings.TrimSpace(string(data)))
	}
	var result chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return chatResponse{}, fmt.Errorf("decode Ollama response: %w", err)
	}
	return result, nil
}

func convertTools(tools []*mcp.Tool) ([]ollamaTool, error) {
	result := make([]ollamaTool, 0, len(tools))
	for _, tool := range tools {
		if tool == nil {
			continue
		}
		var parameters any
		data, err := json.Marshal(tool.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("encode schema for %q: %w", tool.Name, err)
		}
		if err := json.Unmarshal(data, &parameters); err != nil {
			return nil, fmt.Errorf("decode schema for %q: %w", tool.Name, err)
		}
		converted := ollamaTool{Type: "function"}
		converted.Function.Name = tool.Name
		converted.Function.Description = tool.Description
		converted.Function.Parameters = parameters
		result = append(result, converted)
	}
	return result, nil
}
