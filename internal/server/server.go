package server

import (
	"context"
	"fmt"

	"github.com/alifatolahi/ganjoor-mcp/internal/ganjoor"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Server struct {
	API ganjoor.Client
}

func New(api ganjoor.Client) *mcp.Server {
	service := &Server{API: api}
	mcpServer := mcp.NewServer(
		&mcp.Implementation{Name: "farsi-literature-mcp", Version: "0.1.0"},
		nil,
	)
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "get_poem", Description: "Get a Ganjoor poem by numeric ID.",
	}, service.getPoem)
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "get_poet", Description: "Get a Ganjoor poet by numeric ID.",
	}, service.getPoet)
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "get_category", Description: "Get a Ganjoor category by numeric ID.",
	}, service.getCategory)
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "search_poetry", Description: "Search Ganjoor poems by text.",
	}, service.searchPoetry)
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "get_context", Description: "Get verses around a verse in a Ganjoor poem.",
	}, service.getContext)
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name: "get_provenance", Description: "Get source and attribution metadata for a poem.",
	}, service.getProvenance)
	return mcpServer
}

type poemInput struct {
	ID int `json:"id" jsonschema:"Ganjoor poem ID"`
}

type poetInput struct {
	ID int `json:"id" jsonschema:"Ganjoor poet ID"`
}

type categoryInput struct {
	ID int `json:"id" jsonschema:"Ganjoor category ID"`
}

type searchInput struct {
	Query  string `json:"query" jsonschema:"Text to search for"`
	PoetID int    `json:"poet_id,omitempty" jsonschema:"Optional Ganjoor poet ID"`
	Limit  int    `json:"limit,omitempty" jsonschema:"Maximum results, default 20"`
}

type searchOutput struct {
	Results []ganjoor.Poem `json:"results"`
}

type contextInput struct {
	PoemID     int `json:"poem_id" jsonschema:"Ganjoor poem ID"`
	VerseOrder int `json:"verse_order,omitempty" jsonschema:"1-based verse order; omit for the full poem"`
	Before     int `json:"before,omitempty" jsonschema:"Verses before the selected verse"`
	After      int `json:"after,omitempty" jsonschema:"Verses after the selected verse"`
}

func (s *Server) getPoem(ctx context.Context, _ *mcp.CallToolRequest, input poemInput) (*mcp.CallToolResult, ganjoor.Poem, error) {
	poem, err := s.API.GetPoem(ctx, input.ID)
	return nil, poem, err
}

func (s *Server) getPoet(ctx context.Context, _ *mcp.CallToolRequest, input poetInput) (*mcp.CallToolResult, ganjoor.Poet, error) {
	poet, err := s.API.GetPoet(ctx, input.ID)
	return nil, poet, err
}

func (s *Server) getCategory(ctx context.Context, _ *mcp.CallToolRequest, input categoryInput) (*mcp.CallToolResult, ganjoor.Category, error) {
	category, err := s.API.GetCategory(ctx, input.ID)
	return nil, category, err
}

func (s *Server) searchPoetry(ctx context.Context, _ *mcp.CallToolRequest, input searchInput) (*mcp.CallToolResult, searchOutput, error) {
	limit := input.Limit
	if limit == 0 {
		limit = 20
	}
	if limit < 1 || limit > 1000 {
		return nil, searchOutput{}, fmt.Errorf("limit must be between 1 and 1000")
	}
	poems, err := s.API.SearchPoems(ctx, input.Query, input.PoetID, 1, limit)
	return nil, searchOutput{Results: poems}, err
}

func (s *Server) getContext(ctx context.Context, _ *mcp.CallToolRequest, input contextInput) (*mcp.CallToolResult, ganjoor.Context, error) {
	result, err := s.API.GetContext(ctx, input.PoemID, input.VerseOrder, input.Before, input.After)
	return nil, result, err
}

func (s *Server) getProvenance(ctx context.Context, _ *mcp.CallToolRequest, input poemInput) (*mcp.CallToolResult, ganjoor.Provenance, error) {
	result, err := s.API.GetProvenance(ctx, input.ID)
	return nil, result, err
}
