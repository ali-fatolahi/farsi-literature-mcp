package main

import (
	"context"
	"log"

	"github.com/alifatolahi/ganjoor-mcp/internal/ganjoor"
	"github.com/alifatolahi/ganjoor-mcp/internal/server"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	api := ganjoor.NewClient()
	if err := server.New(api).Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
