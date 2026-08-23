package main

import "testing"

func TestDefaultMCPCommandUsesGoRun(t *testing.T) {
	args := defaultMCPArgs()
	if len(args) != 2 || args[0] != "run" || args[1] != "./cmd/ganjoor-mcp" {
		t.Fatalf("unexpected default MCP command: %v", args)
	}
}
