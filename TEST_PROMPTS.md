# Test Prompts

## Verified poetry search

Use this prompt to test an MCP client, especially after changing tool
definitions, API handling, or retrieval behavior:

```text
Find three Persian classical poetry passages about separation and longing (فراق و دلتنگی).

For each result, provide:
- Poet
- Work/title
- Exact original Persian text
- Ganjoor URL
- A one-sentence explanation of why it matches

Do not invent quotations, attribution, or URLs. If you cannot verify a result, say so clearly.
```

## Expected behavior

The client should call `search_poetry` before making literary evidence claims.
Results should use exact text returned by Ganjoor, preserve attribution, and
include URLs derived from the returned poem records. Unverified or missing
results should be stated explicitly rather than filled with generated text.

## Evaluation note

The separation-and-longing prompt is a useful smoke test, but it does not yet
demonstrate a strong advantage over Claude's built-in response because the
explicit Persian terms (`فراق` and `دلتنگی`) make direct keyword search easy.
Add harder evaluation prompts later, such as thematic paraphrases, spelling
and spacing variants, ambiguous attribution, context-dependent requests, and
queries requiring verified provenance. Consider expanding the server's
retrieval features before treating this prompt as evidence of distinct MCP
value.

The initial comparisons also suggest a ranking tradeoff: Claude without MCP
often leads with popular or canonical sources, while the current MCP search
may surface less-known resources through flatter ranking. Future evaluation
should measure both relevance and source diversity, and consider whether
popularity or canonicality should be an explicit ranking signal.
