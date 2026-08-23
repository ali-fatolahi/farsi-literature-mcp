# Ganjoor MCP

An MCP server and AI-agent interface for searching Persian classical literature by meaning, theme, and context, not only by exact keywords.

The initial target is [Ganjoor](https://ganjoor.net), a large public collection of Persian poetry. This project is in early implementation.

## Project Goals

- Make thematic and contextual search useful for Persian poetry.
- Preserve exact quotations, attribution, provenance, and source links.
- Support Persian and English natural-language queries.
- Expose focused read-only tools to MCP-compatible AI clients.
- Keep retrieval evidence separate from model-generated explanation.

## Repository Guide

| File | Purpose |
| --- | --- |
| [`README.md`](README.md) | Project overview, setup, usage, and links |
| [`DESIGN.md`](DESIGN.md) | Architecture, design decisions, tradeoffs, and open questions |
| [`PROGRESS.md`](PROGRESS.md) | Dated progress log and next implementation steps |

## Current Status

The project is in early implementation. The API client, snapshot downloader,
and initial stdio MCP server are available.

## Research Summary

Ganjoor provides browsing by poet and work, keyword search, poem similarity, annotations, and metadata such as language, form, metre, and word-frequency statistics. Its public data ecosystem includes:

- [GanjoorService](https://github.com/ganjoor/GanjoorService), the official open-source site/backend code and API reference.
- [ganjoor-data](https://github.com/ganjoor/ganjoor-data), the official public export of poets, categories, and poems.
- [persian-poetry-ai-agent-plugin](https://github.com/ganjoor/persian-poetry-ai-agent-plugin), an existing JSON-to-Markdown, QMD, and MCP experiment for this problem.

The latest research notes, corpus figures, attribution considerations, and architectural rationale are maintained in [`DESIGN.md`](DESIGN.md).

## Setup

Requirements: Go 1.23 or newer for the MCP SDK integration.

```sh
git clone https://github.com/ali-fatolahi/farsi-literature-mcp.git
cd farsi-literature-mcp
go mod download
```

## Code Structure

| Path | Purpose |
| --- | --- |
| `cmd/ganjoor-fetch` | Command-line downloader for pinned `ganjoor-data` snapshots |
| `cmd/ganjoor-mcp` | Stdio MCP server entry point |
| `cmd/ollama-mcp` | Interactive Ollama-to-MCP bridge |
| `internal/ganjoor` | Ganjoor API client, response models, and upstream error handling |
| `internal/bridge` | MCP tool discovery, Ollama tool translation, and tool-call loop |
| `internal/server` | MCP tool registration and handlers |
| `internal/snapshot` | Snapshot manifest validation and recursive export downloading |
| `.github/workflows/ci.yml` | Formatting, vet, and test checks run by GitHub Actions |
| `DESIGN.md` | Architecture and design decisions |
| `PROGRESS.md` | Implementation status and next steps |

## Development

Run the same checks used by CI:

```sh
gofmt -l .
CGO_ENABLED=0 go vet ./...
CGO_ENABLED=0 go test ./...
```

The snapshot downloader requires an exact upstream commit and is intended for
the future local-indexed mode:

```sh
go run ./cmd/ganjoor-fetch \
  -commit f5fdf5152fce7efaf287b2115195efe0a9505b14 \
  -output data/snapshots/2026-08-16
```

The downloader may take a long time and can encounter upstream rate limits.
The API client is used by the MCP server and is also available as a library
package.

Start the current stdio MCP server with:

```sh
go run ./cmd/ganjoor-mcp
```

It currently exposes read-only poem, poet, category, search, context, and
provenance tools for MCP-compatible clients.

To use those tools with Ollama's local model API, start the bridge while the
Ollama macOS app is running and the selected model is available:

```sh
ollama pull qwen3
go run ./cmd/ollama-mcp -model qwen3
```

Type one prompt per line. The bridge starts the MCP server automatically,
discovers its tools, forwards Ollama tool calls, and prints the final response.
Progress logs are written to stderr so they remain visible without mixing into
the response text. The Ollama app itself does not need a separate MCP
configuration.

## Planned MCP Tools

- `search_poetry(query, poet?, form?, language?, limit?)`
- `find_similar_passage(text, poet?, limit?)`
- `get_poem(poem_id)`
- `get_poet(poet_id_or_slug)`
- `get_context(poem_id, verse_id?, before?, after?)`
- `get_provenance(poem_id)`
- `search_annotations(query, poem_id?, limit?)`

All search results should include stable identifiers and a link to the relevant Ganjoor page.

## Contributing to the Direction

Useful early contributions include sample thematic queries with expected relevant poems, notes about Persian normalization, retrieval evaluation examples, feedback on MCP tool boundaries, and verification of attribution requirements.

The project is published as a GitHub repository with automated formatting,
vetting, and test checks.

See [`PROGRESS.md`](PROGRESS.md) for the current next step.

## License

Not decided yet. The code license will be chosen separately from the rights and attribution conditions of Ganjoor's data, transcriptions, annotations, recordings, and generated artifacts.
