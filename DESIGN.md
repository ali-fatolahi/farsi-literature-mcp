# Design Document

## Purpose

This document records the design of Ganjoor MCP: a local retrieval layer and MCP interface for finding relevant Persian classical literature by exact text, similarity, theme, and context.

The system should retrieve evidence from the Ganjoor corpus. An optional AI agent may plan searches and explain results, but it must not silently replace source text with generated interpretation.

## Problem Statement

Keyword search is useful for known words and remembered lines, but it is a poor interface for questions such as:

- Which poems express separation without using the expected keyword?
- How do Hafez and Saadi describe mortality and the passing of time?
- Where is spring used as a metaphor for renewal?

The system therefore needs lexical retrieval, semantic retrieval, metadata filters, context expansion, and provenance-aware responses.

## Research Findings

Research conducted on 2026-08-19:

- Ganjoor supports browsing by poet and work, collection search, poem similarity, annotations, and metadata views.
- Ganjoor's statistics page reported approximately 1.49 million verses, 26.1 million counted words, and 337,909 unique word forms. These are changing site statistics, not fixed requirements.
- The official public export reported 234 poets and 132,591 poems generated on 2026-08-16.
- The export includes public poetry content and excludes account-linked data such as comments, bookmarks, reading history, and edit history.
- `GanjoorService` is the official open-source ASP.NET Core backend/frontend project and documents the public API at [api.ganjoor.net](https://api.ganjoor.net).
- `ganjoor-data` provides `manifest.json`, JSON content, metadata files, and documented HTTPS access. It is preferable to scraping live HTML for indexing.
- Ganjoor content comes from multiple sources. Source and attribution metadata must remain attached to indexed records.
- The official organization has an existing [Persian poetry AI-agent plugin](https://github.com/ganjoor/persian-poetry-ai-agent-plugin) using Markdown conversion, QMD indexes, multilingual embeddings, and MCP. This project should reuse compatible ideas or contribute upstream rather than duplicate functionality without a clear reason.

### Upstream export schema inspected on 2026-08-22

The inspected snapshot is commit
[`f5fdf5152fce7efaf287b2115195efe0a9505b14`](https://github.com/ganjoor/ganjoor-data/commit/f5fdf5152fce7efaf287b2115195efe0a9505b14),
generated at `2026-08-16T12:20:22.3558565Z`. It reports schema version 1,
234 poets, and 132,591 poems.

- `manifest.json` is the source of truth for the schema version, generation
  time, counts, ID-index shard size, URL templates, and poet IDs/slugs.
- Poet records at `poets/{poetSlug}/poet.json` contain a numeric `Id`,
  `Name`, `Nickname`, biography and place/date metadata, plus `FullUrl`.
- Category records at `poets/{poetSlug}/{catPath}/_cat.json` contain `Id`,
  `PoetId`, `ParentId`, `Title`, `FullUrl`, ordered `ChildCats`, and poem
  summaries with numeric IDs, titles, and URL paths.
- Poem records at `poets/{poetSlug}/{catPath}/{poemSlug}.json` contain `Id`,
  `CatId`, title fields, `FullUrl`, source fields, optional summary and metre,
  sections, and ordered verses. Verses provide `VOrder`, `Position`, `Text`,
  `CoupletIndex`, and `SectionIndex1`; sections provide indexes, type,
  verse type, plain/HTML text, and couplet count.
- Numeric poet IDs resolve through `index/poets-by-id.json`. Category and poem
  IDs resolve through sharded indexes whose bucket is integer division by the
  manifest's `IdIndexShardSize` (currently 2,000). Index entries provide
  Ganjoor URL paths, which map directly to the content URL templates.
- The export has no static search endpoint and excludes account-linked data.
  It is therefore suitable as the reproducible long-term ingestion source; the
  API is the initial MVP source while snapshot ingestion is deferred.

## Architecture

```text
Ganjoor public export / API
            |
            v
  pinned raw snapshot and manifest
            |
            v
 Persian normalization + metadata extraction
            |
            +--> lexical index (BM25 / SQLite FTS / QMD)
            |
            +--> multilingual vector index
            |
            v
 hybrid retrieval + provenance-preserving result model
            |
            v
 MCP tools --> compatible AI agents and clients
```

### MVP with Ollama

For the MVP, use Ollama as the language-model runtime and keep the MCP server
as the only domain-specific service. A thin local chat host is still required:
it maintains the conversation, advertises the MCP tools to the model, forwards
tool calls to the MCP server, and returns tool results to Ollama for the final
response.

```text
human-friendly query
          |
          v
local chat host <--> Ollama API
          |
          v
MCP client <--> Ganjoor MCP server
```

This avoids building a separate planning agent for the first release. Ollama's
model performs the tool-selection loop, while the MCP server owns retrieval,
source text, identifiers, and provenance. The host should enforce a small
maximum tool-call loop and preserve structured tool results in the prompt.

`ollama chat` by itself is not an MCP client. It can be used as the model
backend, but the host or an MCP-capable chat application must provide the
bridge between Ollama's tool-calling API and the MCP server.

The repository includes that bridge as `cmd/ollama-mcp`. It starts
`cmd/ganjoor-mcp` as a child process over MCP stdio, calls `tools/list`, maps
the discovered schemas to Ollama's `tools` JSON field, and forwards each
Ollama `tool_calls` entry to MCP `tools/call`. Tool results are returned to
Ollama as JSON-encoded `tool` messages. The bridge uses a bounded tool-call
loop and reads one user prompt per input line.

### Bridge debugging and evaluation

The first end-to-end run demonstrated that MCP transport initialization and
tool invocation work, but model behavior and evidence flow are not yet
validated. A relevant user request produced an unrelated answer in one run,
and another run generated an unsuitable transliterated search term. Before
changing ranking or adding semantic retrieval, debug the complete loop:

1. Record the exact tool name and arguments selected by Ollama.
2. Record a compact, privacy-conscious summary of each MCP result, including
   result count, IDs, and error state.
3. Add an opt-in trace containing the Ollama message roles and tool-call
   payloads, with response text truncation.
4. Exercise each MCP tool deterministically through an MCP client test.
5. Run the same Persian terms directly against the Ganjoor API and through MCP
   to isolate API, adapter, schema, and model issues.
6. Use a constrained system prompt requiring tool use for claims about poems,
   exact quotations, attribution, and URLs.

This separates transport/protocol failures from model tool-selection failures
and retrieval-quality failures.

### Implementation direction

The MVP will be implemented as a standalone Go server backed by the Ganjoor
API. The pinned local export and downloader remain the long-term reliability
path, but are deferred until resumability, rate limits, and first-time setup
cost are designed properly. The API adapter must isolate upstream failures
with timeouts, bounded retries, rate limiting, and explicit errors.

The initial component choices are:

- Go 1.23 or newer.
- The official `github.com/modelcontextprotocol/go-sdk/mcp` package for the
  MCP server and client interfaces.
- `modernc.org/sqlite` pinned to the v1.34 line for a pure-Go
  `database/sql` driver and SQLite FTS5 lexical indexing without a CGO
  requirement.
- Ollama's HTTP API for local generation and, after evaluation, embeddings.

The initial API contract inspection found `GET /api/ganjoor/poem/{id}` for
complete poem retrieval, `GET /api/ganjoor/poet/{id}` for poet lookup,
`GET /api/ganjoor/cat/{id}` for category metadata, and
`GET /api/ganjoor/poems/search` for lexical search. The Go client begins with
poem retrieval and deliberately disables unrelated expansions such as
recitations, images, songs, comments, navigation, and related poems.

The SDK dependency requires Go 1.23 or newer. Vector storage and embedding
model selection remain intentionally deferred until the lexical baseline is
measured.

## Design Decisions

### Use the public export as the long-term indexing source

**Decision:** Prefer the official `ganjoor-data` export and documented API over crawling Ganjoor HTML pages.

**Reasoning:** The export is structured, versionable, and explicitly excludes private/account-linked data. A pinned snapshot makes an index reproducible and reduces load on the live website.

**Consequence:** The project needs an ingestion and update process, and live site changes will not appear until the next snapshot is indexed. This is
deferred from the MVP because a full unauthenticated export download hit HTTP
429 responses after 39,000 files in 6m39s.

**Schema outcome:** Store the upstream commit SHA and manifest generation
timestamp with each local snapshot. Use numeric IDs as primary keys, retain
the upstream URL path and construct the canonical page URL as
`https://ganjoor.net` plus `FullUrl`. Treat `Sections` and `Verses` as the
source for passage boundaries rather than reparsing `HtmlText`.

### Use Go for the standalone server

**Decision:** Implement the ingestion, local retrieval, and MCP server in Go.

**Reasoning:** Go provides a small deployable binary and straightforward
concurrency and HTTP support, while keeping the domain logic independent of
the language used by Ganjoor's own backend. The existing Python plugin remains
useful as a source of compatible retrieval ideas, not as an implementation
constraint.

**Consequence:** The project still needs to implement the API-backed MVP before
local indexing. Vector retrieval and embedding-model selection can follow the
lexical baseline rather than being required before the first runnable version.

### Use the official MCP SDK and SQLite FTS5

**Decision:** Use the official Go MCP SDK and `modernc.org/sqlite` with FTS5
when local indexing is introduced.

**Reasoning:** The official SDK tracks the MCP specification and supports both
server and client roles needed by the planned thin host. A pure-Go SQLite
driver keeps the first binary free of CGO/compiler requirements while
providing a mature local database and full-text search baseline.

**Compatibility:** Current SDK releases require Go 1.23 or newer. Pin
dependencies rather than tracking `main`; revisit versions during module
creation if the installed toolchain or MCP protocol target changes.

### Keep original and normalized text

**Decision:** Store original Persian text for display and citation alongside a normalized representation for retrieval.

**Reasoning:** Persian and Arabic characters, half-spaces, diacritics, spacing, and spelling variants can prevent lexical matches. Normalization improves recall but can erase distinctions meaningful to readers.

**Consequence:** Search indexes use normalized text, while responses quote only the original stored text.

### Use hybrid retrieval

**Decision:** Combine lexical retrieval with multilingual vector retrieval.

**Reasoning:** Embeddings help with themes and paraphrases. Lexical retrieval remains stronger for exact quotations, names, rare words, and orthographic variants.

**Consequence:** Ranking must combine scores from different systems and expose whether a result is exact, lexical, semantic, or hybrid.

### Index passages and poems

**Decision:** Index both complete poems and smaller passage windows.

**Reasoning:** A couplet gives precise evidence, while adjacent verses often provide the context needed to judge a thematic result.

**Consequence:** Every passage must point back to a stable poem ID and verse range. Results must be deduplicated when multiple windows from one poem match.

### Make provenance part of the result contract

**Decision:** Every result includes poet, work, category, poem ID, verse IDs or range, source metadata when available, and a Ganjoor URL.

**Reasoning:** Literary attribution and textual variants are not always certain. Users need to inspect the original record and distinguish corpus facts from interpretation.

**Consequence:** No result formatter may discard identifiers or source links.

### Expose focused read-only MCP tools

**Decision:** Provide composable search, lookup, context, and provenance tools instead of one unrestricted question-answer tool.

**Reasoning:** Small tools are easier to test, authorize, explain, and use across different MCP clients. They keep evidence retrieval separate from agent reasoning.

**Consequence:** An agent may orchestrate several calls, but the server remains useful without a particular LLM provider.

### Keep the MVP host thin

**Decision:** Start with a minimal Ollama-backed MCP chat host rather than a
separate agent framework.

**Reasoning:** The model can choose among focused read-only tools without an
additional planner. This keeps the first implementation small and makes the
MCP server testable with other clients later.

**Consequence:** The host is responsible for MCP transport, Ollama tool-call
translation, conversation state, and loop limits. It should not contain
Ganjoor retrieval or attribution logic.

### Use Ollama's native tool-calling API for the host

**Decision:** Implement the first host as a small local CLI bridge using
Ollama's `/api/chat` endpoint rather than modifying the Ollama macOS app.

**Reasoning:** Ollama accepts function-tool schemas and returns tool calls,
while MCP provides discovery and execution. Translating between those two
contracts keeps the bridge small and works with the Ollama app's local runtime
without requiring an app-specific plugin mechanism.

**Consequence:** The bridge is currently a terminal interface. An Ollama
desktop UI integration would require a separate MCP-capable frontend or a
future host with its own UI.

## Retrieval Pipeline

1. Download and verify a pinned upstream snapshot.
2. Parse the manifest and resolve poet, category, poem, and verse records.
3. Normalize text while retaining original text and metadata.
4. Generate poem-level and passage-level documents.
5. Build a lexical index.
6. Build a multilingual vector index after benchmarking candidate embedding models.
7. Retrieve candidates from both indexes and combine or rerank them.
8. Expand the winning passage with nearby verses when requested.
9. Return evidence, metadata, scores, and source links.

## MCP Contract Sketch

Initial tools: `search_poetry`, `find_similar_passage`, `get_poem`, `get_poet`, `get_context`, `get_provenance`, and `search_annotations`.

Responses should contain structured fields for `text`, `match_type`, `score`, `poet`, `work`, `poem_id`, `verse_ids`, `source`, and `url`.

## Evaluation Plan

Create a small relevance-labeled set containing Persian exact-line queries with spelling and spacing variations, Persian thematic queries, English thematic queries, poet/form constrained queries, and ambiguous queries requiring clarification.

Track recall@k, precision@k, mean reciprocal rank, latency, duplicate rate, and attribution/URL correctness. Exact-line performance must not regress when semantic retrieval is added.

## Boundaries and Risks

- Model-generated summaries and keywords are derived data and must not replace primary text.
- A poem's presence in the corpus is not proof of uncontested authorship or textual authenticity.
- Public-domain status of old poems does not automatically settle rights around transcriptions, compilations, annotations, recordings, or generated summaries.
- Do not ingest private or account-linked user data.
- Rate-limit remote access and make the local index the normal serving path.
- The project is not affiliated with Ganjoor unless that relationship is explicitly established.

## Open Decisions

- Extend the existing Persian poetry AI-agent plugin, contribute improvements upstream, or build an independent implementation?
- Which Persian-capable embedding model performs best on the evaluation set?
- Should the default retrieval unit be a couplet, a poem section, or an adaptive context window?
- Should annotations be searchable by default or only through a separate tool?
- Is local-only operation required for privacy and reproducibility?
- How should duplicate poems, disputed attribution, and multiple textual versions be represented?

## References

- [Ganjoor](https://ganjoor.net)
- [Ganjoor about page](https://ganjoor.net/about)
- [Ganjoor search](https://ganjoor.net/search)
- [Ganjoor statistics](https://ganjoor.net/vazn)
- [Ganjoor sources](https://ganjoor.net/sources)
- [GanjoorService](https://github.com/ganjoor/GanjoorService)
- [ganjoor-data](https://github.com/ganjoor/ganjoor-data)
- [Persian poetry AI-agent plugin](https://github.com/ganjoor/persian-poetry-ai-agent-plugin)
