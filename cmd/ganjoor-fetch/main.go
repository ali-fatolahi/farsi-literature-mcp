package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/alifatolahi/ganjoor-mcp/internal/snapshot"
)

func main() {
	commit := flag.String("commit", "", "ganjoor-data commit SHA to download")
	output := flag.String("output", "data/snapshots/current", "snapshot output directory")
	baseURL := flag.String("base-url", "https://raw.githubusercontent.com/ganjoor/ganjoor-data", "upstream raw content base URL")
	flag.Parse()

	if *commit == "" {
		log.Fatal("-commit is required; use an exact ganjoor-data commit SHA")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	metadata, err := (snapshot.Downloader{
		BaseURL: *baseURL,
		Commit:  *commit,
		Output:  *output,
	}).Download(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("downloaded %d files for commit %s into %s\n", metadata.Files, metadata.Commit, *output)
}
