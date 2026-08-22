# Project Progress

This file records implementation progress. Design rationale belongs in [`DESIGN.md`](DESIGN.md); general project information belongs in [`README.md`](README.md).

## Status

**Current phase:** research and design  
**Last updated:** 2026-08-22  
**Runnable server:** not yet available

## Completed

- [x] Defined the problem: semantic and contextual search over Ganjoor's Persian literature.
- [x] Researched Ganjoor's site capabilities, public API, source information, and statistics pages.
- [x] Confirmed the official `ganjoor-data` public export.
- [x] Confirmed that the export excludes account-linked user data.
- [x] Identified `GanjoorService` as the official open-source backend/frontend repository.
- [x] Identified the existing Persian poetry AI-agent plugin and its QMD/MCP approach.
- [x] Recorded the initial architecture and design decisions in [`DESIGN.md`](DESIGN.md).
- [x] Restructured [`README.md`](README.md) as a project guide.
- [x] Chose a thin Ollama-backed MCP chat host as the MVP integration path.
- [x] Inspected the pinned `ganjoor-data` manifest, API documentation, and representative poet, category, poem, and ID-index records.
- [x] Chose Go as the MVP implementation direction.
- [x] Selected the official Go MCP SDK and a pure-Go SQLite driver with FTS5 support for the lexical baseline.
- [x] Implemented a reproducible snapshot downloader with manifest validation and recursive content discovery.
- [x] Deferred full snapshot ingestion for the MVP after measuring unauthenticated download rate limits; use the Ganjoor API initially.
- [x] Inspected the published API contract and added a Go client foundation with `get_poem`, timeouts via contexts, bounded retries, and typed upstream errors.
- [x] Connected the local project to `ali-fatolahi/farsi-literature-mcp` and added GitHub Actions CI for formatting, vet, and tests.

## Next Steps

### Step 0: establish the GitHub project and CI

- The local repository is initialized on `main` with `origin` set to
  `ali-fatolahi/farsi-literature-mcp`.
- GitHub Actions workflow added for formatting checks, `go vet`, and tests.
- CI runs with CGO disabled for the current pure-Go code path.
- Push the initial commit after review.
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

- Complete the API adapter for poet, context, and search operations.
- Add client-side rate limiting in addition to bounded retries and clear
  upstream errors.
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

### Step 6: add semantic and hybrid retrieval

- Build a relevance-labeled query set.
- Benchmark multilingual embedding candidates in Persian and English.
- Combine lexical and vector candidates without degrading exact-line search.

### Step 7: expose MCP tools

- Implement read-only search, lookup, context, and provenance tools.
- Validate structured responses and source links.
- Test with at least one MCP client.

## Decision Log

| Date | Decision | Reference |
| --- | --- | --- |
| 2026-08-19 | Keep README focused on orientation and setup; move architecture to a design document and execution state to this file. | Repository documentation restructure |
| 2026-08-19 | Prefer the official public export over HTML scraping for the indexing source. | [`DESIGN.md`](DESIGN.md) |
| 2026-08-19 | Use hybrid lexical and semantic retrieval. | [`DESIGN.md`](DESIGN.md) |
| 2026-08-19 | Preserve original text and provenance in every result. | [`DESIGN.md`](DESIGN.md) |
| 2026-08-20 | Use Ollama plus a thin MCP chat host for the MVP; no separate agent framework initially. | [`DESIGN.md`](DESIGN.md) |

## Blockers and Notes

- The Go MCP SDK and index technology still need to be selected.
- No upstream snapshot has been downloaded into this repository.
- Licensing and redistribution requirements need review before publishing generated indexes or summaries.
