# Evaluation comparison: internal oppression and foreign intervention

## Runs compared

- Without MCP: `2026-08-26-foreign-intervention-without-mcp.md`
- With MCP: `2026-08-26-foreign-intervention-with-mcp.md`
- Compared at (date known; time not recorded): 2026-08-26

## Result comparison

| Criterion | Without MCP | With MCP |
| --- | --- | --- |
| Fit to the exact claim | More direct, via the Jamshid/Zahhak interpretation | More indirect, focused on vulnerability to external enemies |
| Exact quotations | Not provided | Several quotations provided |
| Verified Ganjoor URLs | None | None |
| Stable poem identifiers | None | None |
| Primary-source verification | Not demonstrated | Not demonstrated |
| Immediate usefulness | Clear hypothesis needing verification | Potentially useful leads needing verification |

## Distinct MCP value

The MCP response produced multiple candidate passages and quotations, but it
did not provide the URLs or IDs needed to verify them. In this run, the
without-MCP response gave a more direct interpretation of the requested
scenario, while the MCP response supplied more textual leads.

Across this and the metaphor query, Claude without MCP appeared to prioritize
more popular or canonical resources. The MCP search behaved more like a flat
corpus search and surfaced less-known resources, which may increase coverage
but can make the first results feel less immediately relevant.

## Next tests and implementation targets

- Capture tool names, arguments, and compact returned records.
- Require every claimed passage to include a poem ID and Ganjoor URL.
- Distinguish “people invite or submit to a foreign ruler” from the broader
  claim that domestic oppression merely enables foreign invasion.
- Verify quotations and attribution by fetching the referenced poems before
  presenting them as evidence.
