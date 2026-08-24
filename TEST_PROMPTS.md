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
