# Project Progress

This file records implementation progress. Design rationale belongs in [`DESIGN.md`](DESIGN.md); general project information belongs in [`README.md`](README.md).

## Status

**Current phase:** API-backed MVP evaluation
**Last updated:** 2026-08-26
**Runnable server:** initial stdio server available

## Completed

- [x] Defined the problem: semantic and contextual search over Ganjoor's Persian literature.
- [x] Researched Ganjoor's site capabilities, public API, source information, and statistics pages.
- [x] Confirmed the official `ganjoor-data` public export.
- [x] Confirmed that the export excludes account-linked user data.
- [x] Identified `GanjoorService` as the official open-source backend/frontend repository.
- [x] Identified the existing Persian poetry AI-agent plugin and its QMD/MCP approach.
- [x] Recorded the initial architecture and design decisions in [`DESIGN.md`](DESIGN.md).
- [x] Restructured [`README.md`](README.md) as a project guide.
- [x] Chose Claude Desktop as the initial MCP client and removed the custom Ollama bridge from the MVP path.
- [x] Inspected the pinned `ganjoor-data` manifest, API documentation, and representative poet, category, poem, and ID-index records.
- [x] Chose Go as the MVP implementation direction.
- [x] Selected the official Go MCP SDK and a pure-Go SQLite driver with FTS5 support for the lexical baseline.
- [x] Implemented a reproducible snapshot downloader with manifest validation and recursive content discovery.
- [x] Deferred full snapshot ingestion for the MVP after measuring unauthenticated download rate limits; use the Ganjoor API initially.
- [x] Inspected the published API contract and added a Go client with poem, poet, category, lexical search, context, and provenance methods, context-aware requests, bounded retries, rate limiting, and typed upstream errors.
- [x] Wired the API client into an initial stdio MCP server with read-only poem, poet, category, search, context, and provenance tools, plus handler tests.
- [x] Verified the server can be configured as a Claude Desktop stdio MCP server.
- [x] Stored the baseline separation-and-longing test prompt in [`TEST_PROMPTS.md`](TEST_PROMPTS.md).
- [x] Ran initial Claude Desktop comparisons with and without the MCP server,
  including explicit-keyword, metaphorical, and political-literary prompts.
- [x] Recorded evaluation results and identified ranking, provenance, and
  tool-result-size gaps in [`evaluations/`](evaluations/).
- [x] Connected the local project to `ali-fatolahi/farsi-literature-mcp` and added GitHub Actions CI for formatting, vet, and tests.

## Next Steps

### Step 0: establish the GitHub project and CI

- The local repository is initialized on `main` with `origin` set to
  `ali-fatolahi/farsi-literature-mcp`.
- GitHub Actions workflow added for formatting checks, `go vet`, and tests.
- CI runs with CGO disabled for the current pure-Go code path.
- Initial commits are pushed to `main`.
- Add a minimal contribution and issue workflow once the repository exists.

### Step 1: inspect the upstream schema

Completed on 2026-08-22. The schema notes and ingestion decision are recorded in [`DESIGN.md`](DESIGN.md).

### Step 2: choose the initial implementation stack

Completed on 2026-08-22. The selected components and compatibility notes are
recorded in [`DESIGN.md`](DESIGN.md).

### Step 3: build a reproducible corpus snapshot

The downloader is implemented but deferred from the MVP. A full unauthenticated
download triggered HTTP 429 responses after 39,000 files in 6m39s, confirming
that snapshot ingestion needs additional design work around resumability,
mirrors, authentication, and rate limits.

- Design resumable and rate-limit-aware snapshot ingestion.
- Keep raw upstream data separate from generated normalized/indexed data.

### Step 4: implement the API-backed MVP

- Expand MCP request/response validation and test the server through Claude
  Desktop or another MCP client.
- Debug the end-to-end tool loop before improving retrieval quality:
  - Log selected tool arguments and compact result summaries.
  - Capture exact MCP tool names, arguments, and returned summaries in an
    opt-in trace mode.
  - Add deterministic MCP client tests for every registered tool.
  - Compare direct Ganjoor API results with results returned through MCP.
  - Add compact search result types containing stable IDs, titles, poet
    metadata, short excerpts, and canonical URLs; reserve full poem text for
    explicit lookup/context calls so MCP responses stay below client limits.
  - Require source links and identifiers in evidence-oriented result
    formatting.
  - Measure popularity/canonicality, relevance, and source diversity
    separately; decide whether ranking should promote canonical sources.
  - Add prompts that test thematic paraphrases, spelling variants, ambiguous
    attribution, and context-dependent requests.
- Keep the local API adapter separate so it can later be replaced by snapshot
  indexes.

### Step 5: implement Persian normalization

- Define handling for Arabic/Persian character variants, diacritics, whitespace, half-spaces, punctuation, and common spelling variants.
- Preserve original text unchanged.
- Add unit tests using real corpus examples and remembered-line queries.

### Step 6: establish the lexical baseline

- Index poems and passage windows.
- Support exact-line search, normalized search, poet filters, and source links.
- Measure baseline retrieval before adding embeddings.

### Step 7: add semantic and hybrid retrieval

- Build a relevance-labeled query set.
- Benchmark multilingual embedding candidates in Persian and English.
- Combine lexical and vector candidates without degrading exact-line search.

### Step 8: extend and evaluate MCP tools

- Improve the existing read-only search, lookup, context, and provenance
  tools using the evaluation findings above.
- Validate structured responses and source links.
- Continue testing with Claude Desktop and recorded paired evaluations.

## Decision Log

| Date | Decision | Reference |
| --- | --- | --- |
| 2026-08-19 | Keep README focused on orientation and setup; move architecture to a design document and execution state to this file. | Repository documentation restructure |
| 2026-08-19 | Prefer the official public export over HTML scraping for the indexing source. | [`DESIGN.md`](DESIGN.md) |
| 2026-08-19 | Use hybrid lexical and semantic retrieval. | [`DESIGN.md`](DESIGN.md) |
| 2026-08-19 | Preserve original text and provenance in every result. | [`DESIGN.md`](DESIGN.md) |
| 2026-08-24 | Use Claude Desktop as the initial MCP client; remove the custom Ollama bridge from the MVP path. | [`DESIGN.md`](DESIGN.md) |
| 2026-08-26 | Treat compact, provenance-rich search results and ranking evaluation as the next API-backed MVP priorities. | [`evaluations/`](evaluations/) |

## Blockers and Notes

- No upstream snapshot has been downloaded into this repository.
- Licensing and redistribution requirements need review before publishing generated indexes or summaries.
