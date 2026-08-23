package bridge

import (
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestConvertToolsUsesOllamaFunctionShape(t *testing.T) {
	tools, err := convertTools([]*mcp.Tool{{
		Name:        "get_poem",
		Description: "Get a poem",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "integer"},
			},
		},
	}})
	if err != nil {
		t.Fatalf("convertTools returned error: %v", err)
	}
	if len(tools) != 1 || tools[0].Type != "function" || tools[0].Function.Name != "get_poem" {
		t.Fatalf("unexpected converted tool: %+v", tools)
	}
}
