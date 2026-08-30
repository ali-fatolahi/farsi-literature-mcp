# Ganjoor MCP

[![CI](https://github.com/ali-fatolahi/farsi-literature-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/ali-fatolahi/farsi-literature-mcp/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

An MCP server and AI-agent interface for searching Persian classical literature by meaning, theme, and context, not only by exact keywords.

GitHub short description:

Experimental open-source POC for an AI-assisted Persian literature hub. Uses Ganjoor as a public, provenance-aware testbed for literary retrieval, evaluation, and MCP tooling.

## Public project status

This repository is an experimental, open-source proof of concept for an AI-assisted Persian literature hub. The current implementation uses Ganjoor as a pragmatic public testbed because it is accessible, well known, and useful for evaluating provenance-aware literary retrieval.

This project is not claiming that a local MCP server is already a universally better replacement for a general-purpose model such as Claude Desktop. The goal is to explore evidence-first retrieval, exact-source quoting, and evaluation methods that make literary search more transparent and more verifiable.

We welcome contributions in three forms:

- code
- ideas, use cases, and design proposals
- test results and prompt evaluations

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
| [`TEST_PROMPTS.md`](TEST_PROMPTS.md) | Reusable prompts for client and retrieval checks |

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

Requirements: Go 1.27 or newer for the MCP SDK integration.

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
| `internal/ganjoor` | Ganjoor API client, response models, and upstream error handling |
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

To use the server with Claude Desktop, build the server and add the binary to
Claude Desktop's MCP configuration file. Use absolute paths:

```json
{
  "mcpServers": {
    "farsi-literature": {
      "command": "/Users/yourname/bin/ganjoor-mcp"
    }
  }
}
```

```sh
mkdir -p ~/bin
go build -o ~/bin/ganjoor-mcp ./cmd/ganjoor-mcp
```

The Claude Desktop configuration file on macOS is usually
`~/Library/Application Support/Claude/claude_desktop_config.json`. Restart
Claude Desktop after saving the configuration. The server currently uses the
Ganjoor API directly; no local corpus download is required.

### Run the MVP in Claude Desktop

1. Run the repository checks and build the binary:

   ```sh
   CGO_ENABLED=0 go test ./...
   mkdir -p ~/bin
   CGO_ENABLED=0 go build -o ~/bin/ganjoor-mcp ./cmd/ganjoor-mcp
   ```

2. Add the `farsi-literature` entry above to
   `~/Library/Application Support/Claude/claude_desktop_config.json`, replacing
   `/Users/yourname` with the actual home directory.
3. Restart Claude Desktop. Approve the server/tool permission request when
   prompted.
4. Open a new chat and confirm that the tools menu lists `search_poetry`,
   `get_poem`, `get_poet`, `get_category`, `get_context`, and
   `get_provenance`.
5. Run a prompt from `TEST_PROMPTS.md`. Require exact quotations, poet/work
   attribution, and Ganjoor URLs; do not treat an unverified answer as a
   successful retrieval.
6. Repeat the same prompt with the MCP server removed from the config, then
   record both responses in `evaluations/`.

If the server fails to connect, inspect
`~/Library/Logs/Claude/mcp-server-farsi-literature.log`. A successful startup
includes `Server started and connected successfully`. The server communicates
over stdio, so diagnostic output must not be written to stdout.

## Planned MCP Tools

- `search_poetry(query, poet?, form?, language?, limit?)`
- `find_similar_passage(text, poet?, limit?)`
- `get_poem(poem_id)`
- `get_poet(poet_id_or_slug)`
- `get_context(poem_id, verse_id?, before?, after?)`
- `get_provenance(poem_id)`
- `search_annotations(query, poem_id?, limit?)`

All search results should include stable identifiers and a link to the relevant Ganjoor page.

## Contributing to the project

We welcome design proposals, bug reports, code contributions, and evaluation notes. See [`CONTRIBUTING.md`](CONTRIBUTING.md) for the current contribution workflow, required review process, and testing expectations.

The project is intended to be community-friendly and respectful. See [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md) for the expected behavior.

Community submissions can be added under `community/` in one of three categories:

- `community/ideas/<github-username>/<timestamp>/idea.md`
- `community/design-proposals/<github-username>/<timestamp>/proposal.md`
- `community/test-results/<github-username>/<timestamp>/result.md`

The file should be a plain Markdown document and the username/timestamp must follow the repository's naming convention. CI validates the directory layout via `scripts/check_contrib_paths.py`.

Useful early contributions include sample thematic queries with expected relevant poems, notes about Persian normalization, retrieval evaluation examples, feedback on MCP tool boundaries, and verification of attribution requirements.

The project is published as a GitHub repository with automated formatting,
vetting, and test checks.

See [`PROGRESS.md`](PROGRESS.md) for the current next step.

## License

This project is licensed under the MIT License. See [`LICENSE`](LICENSE) for the full text. This is separate from the licensing and attribution requirements for Ganjoor's data, transcriptions, annotations, recordings, and generated artifacts.
