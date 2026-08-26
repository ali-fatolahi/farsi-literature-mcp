# Evaluation comparison: دل به کبوتر

## Runs compared

- Without MCP: `2026-08-26-dil-kabutar-without-mcp.md`
- With MCP: `2026-08-26-dil-kabutar-with-mcp.md`
- Compared at (date known; time not recorded): 2026-08-26

## Result comparison

| Criterion | Without MCP | With MCP |
| --- | --- | --- |
| Metaphorical relevance | Suggested two candidates | Suggested four candidates |
| Exact quotations | Not provided | Provided |
| Verified Ganjoor URLs | None | None |
| Stable poem identifiers | None | None |
| Tool-result size | Not applicable | Exceeded Claude's reported 1 MB limit |
| Immediate usefulness | Low without follow-up | Better evidence, but blocked by oversized tool output |

## Distinct MCP value

The MCP run attempted evidence retrieval for a metaphorical query and returned
more concrete candidate quotations than the no-MCP run. However, the current
tool contract did not deliver compact, verifiable results: URLs and IDs were
missing from the final answer, and the raw tool result exceeded the client
size limit.

Claude without MCP appeared to prioritize more popular or canonical resources,
while the MCP search surfaced a flatter mix that included less-known poets.
This may improve breadth but can reduce perceived relevance when users expect
the most recognized examples first.

## Next tests and implementation targets

- Capture the exact MCP tool name and arguments.
- Return compact search-result records containing only ID, title, poet,
  matching text or short excerpt, and URL.
- Fetch full poem text only through an explicit `get_poem` call.
- Repeat this prompt after compacting search output and verify every quote
  against its returned URL.
