# Ganjoor MCP

An MCP server and AI-agent interface for searching Persian classical literature by meaning, theme, and context, not only by exact keywords.

The initial target is [Ganjoor](https://ganjoor.net), a large public collection of Persian poetry. This project is currently in the research and design stage.

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

Research and design. No runnable server has been implemented yet.

## Research Summary

Ganjoor provides browsing by poet and work, keyword search, poem similarity, annotations, and metadata such as language, form, metre, and word-frequency statistics. Its public data ecosystem includes:

- [GanjoorService](https://github.com/ganjoor/GanjoorService), the official open-source site/backend code and API reference.
- [ganjoor-data](https://github.com/ganjoor/ganjoor-data), the official public export of poets, categories, and poems.
- [persian-poetry-ai-agent-plugin](https://github.com/ganjoor/persian-poetry-ai-agent-plugin), an existing JSON-to-Markdown, QMD, and MCP experiment for this problem.

The latest research notes, corpus figures, attribution considerations, and architectural rationale are maintained in [`DESIGN.md`](DESIGN.md).

## Planned Setup

The setup instructions will be expanded when the implementation stack is selected. The planned local workflow is:

```text
1. Clone the project
2. Install Go 1.23+ and the selected dependencies
3. Configure access to the Ganjoor API
4. Start the MCP server
5. Connect an MCP-compatible client
```

The pinned export downloader is retained for the planned local-indexed mode,
but is not required by the MVP.

The snapshot downloader is the first runnable component. The MCP server
entry point will be added after ingestion and lexical indexing are complete.

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

The project will be published as a GitHub repository with automated
formatting, vetting, and test checks before the implementation grows further.

See [`PROGRESS.md`](PROGRESS.md) for the current next step.

## License

Not decided yet. The code license will be chosen separately from the rights and attribution conditions of Ganjoor's data, transcriptions, annotations, recordings, and generated artifacts.
