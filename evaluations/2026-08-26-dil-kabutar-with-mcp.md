# Evaluation: دل به کبوتر with Ganjoor MCP

## Run metadata

- Recorded at (date known; time not recorded): 2026-08-26
- MCP server state: enabled
- Prompt: `چند شعر بیاب که در آن دل به کبوتر تشبیه شده بشاد`

## Claude response

Claude returned examples attributed to Jahan Malek Khatun, Khajoo Kermani,
Salman Savoji, and Naser Khosrow, with Persian quotations and explanations.
The response did not include Ganjoor URLs or stable poem identifiers.

## MCP activity

- Tool result issue: Claude reported that the MCP tool result exceeded the
  apparent 1 MB limit.
- Tools called: Not recorded in the response.
- Notable behavior: The final answer was concise, but the underlying tool
  result was too large for Claude to process reliably.

## Observations

- This prompt tests metaphorical retrieval rather than an explicit search
  term and is more discriminating than the separation prompt.
- The response produced a broader set of candidate poets and quotations.
- Missing URLs and IDs prevent source verification.
- Search results likely need a compact response projection instead of returning
  full poem records and large text fields.
