# Copilot instructions

## Project overview

This repository is a Go 1.23+ MCP server for read-only access to Ganjoor
Persian literature. The current runtime uses the Ganjoor HTTP API directly;
the pinned `ganjoor-data` export and local indexing pipeline are longer-term
work, not required for the current server.

## Build, test, and lint

Run the same checks as GitHub Actions:

```sh
test -z "$(gofmt -l .)"
CGO_ENABLED=0 go vet ./...
CGO_ENABLED=0 go test ./...
```

Run one package or one test selectively:

```sh
CGO_ENABLED=0 go test ./internal/server -run '^TestSearchPoetryHandlerReturnsObjectOutput$'
CGO_ENABLED=0 go test ./internal/ganjoor -run '^TestGetPoemRetriesTransientFailure$'
```

There is no separate linter configured. Keep formatting compatible with
`gofmt`, and preserve the CI requirement that the code works with CGO
disabled.

Run the applications with:

```sh
go run ./cmd/ganjoor-mcp
go run ./cmd/ganjoor-fetch -commit <exact-ganjoor-data-commit-sha> -output data/snapshots/<name>
```

The MCP server communicates over stdio. Do not write diagnostic output to
stdout, because it is the MCP protocol stream; use the standard logger for
fatal startup errors.

## Architecture

The executable in `cmd/ganjoor-mcp` constructs a configured
`internal/ganjoor.Client`, passes it to `internal/server.New`, and runs the
official Go MCP SDK over `mcp.StdioTransport`.

`internal/server` is the MCP boundary. It registers focused read-only tools
(`get_poem`, `get_poet`, `get_category`, `search_poetry`, `get_context`, and
`get_provenance`), validates tool inputs, and returns structured Go values.
Keep model/tool orchestration outside this repository's server; clients such
as Claude Desktop discover and call these tools.

`internal/ganjoor` owns upstream API details, response models, canonical
Ganjoor URLs, provenance mapping, context expansion, typed HTTP errors,
bounded retries, request timeouts via contexts, and rate limiting. Handlers
should call this client rather than constructing API requests themselves.

`internal/snapshot` is the reproducible-export path. It validates a pinned
manifest, maps upstream URL paths into `data/.../raw`, recursively discovers
categories and poems, and downloads content with bounded parallelism and
retry handling. Preserve the raw upstream records and snapshot metadata when
extending this path; do not silently replace them with generated data.

## Repository-specific conventions

- Use numeric Ganjoor IDs as stable identifiers. Preserve upstream `FullURL`
  paths and construct source links as `https://ganjoor.net` plus that path.
- Keep original Persian text unchanged for responses and citations. Future
  normalization is for retrieval/indexing only and must not replace source
  text.
- Preserve attribution and provenance fields (`poem_id`, poet/category IDs,
  source metadata, disputed/multiple-poet status, and URL) in result models.
  Do not discard identifiers or source links in formatters or new tools.
- Treat API input errors as explicit errors: IDs must be positive, search
  terms must be non-empty, limits must be bounded, and context sizes must not
  be negative. Follow existing error-return patterns rather than silently
  accepting invalid values.
- Use request contexts for all upstream calls. Keep retries bounded and only
  retry transient transport failures, HTTP 429, and HTTP 5xx responses.
  Continue to respect the client's rate limiter.
- Keep API models and upstream endpoint behavior in `internal/ganjoor`;
  handlers in `internal/server` should remain thin adapters with JSON-schema
  tagged input/output types.
- MCP tool outputs are structured objects. Avoid changing fields to
  `null`/omitted values unintentionally; existing tests verify sparse output
  behavior.
- Snapshot paths originate from upstream URL paths and must remain traversal
  safe. Use the existing `contentPath` validation and exact commit SHA
  requirement when extending downloads.
- Tests use Go's standard `testing` package and `httptest` servers to exercise
  API behavior without live Ganjoor calls. Follow this pattern for new HTTP
  client and handler tests.
- Keep the API-backed adapter separable from future local retrieval/indexing.
  Do not couple MCP handlers to snapshot storage or embedding choices.

## Authoritative project references

- `README.md`: setup, commands, current usage, and MCP client configuration.
- `DESIGN.md`: long-term retrieval architecture, provenance requirements,
  normalization policy, and deferred indexing decisions.
- `PROGRESS.md`: current implementation status and next steps.
- `TEST_PROMPTS.md`: end-to-end MCP prompt and evidence-preservation
  expectations.
