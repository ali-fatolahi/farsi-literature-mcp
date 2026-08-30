# Contributing to Ganjoor MCP

This repository is open to design proposals, code changes, issue reports, and evaluation results. We value careful, evidence-based improvements more than rapid feature churn.

## Scope of contributions

We welcome:

- Bug reports with reproduction steps and evidence
- Feature requests and design proposals
- Small and large code changes in Go, tests, or documentation
- Evaluation results comparing Claude Desktop with and without the MCP server
- Test prompts, retrieval notes, and provenance/attribution findings

We do not accept:

- Unverified literary claims presented as factual Ganjoor results
- Source text or annotations copied without attribution
- Changes that silently discard provenance, source URLs, or poem IDs
- Non‑reviewed direct commits to `main`

## Contribution workflow

1. Open an issue before starting a large change or design work.
2. For significant features or architecture changes, open a design proposal issue and describe the tradeoffs.
3. Create a branch from `main`, make the change, and add or update tests.
4. Run the repository checks locally before opening a pull request:

   ```sh
   gofmt -l .
   CGO_ENABLED=0 go vet ./...
   CGO_ENABLED=0 go test ./...
   ```

5. Open a pull request with a clear summary, testing steps, evidence, and any documentation updates.
6. Await review and approval before merge.

## Review and approval requirements

For this project, code changes must meet all of the following:

- Pass the required CI checks (`gofmt`, `go vet`, and `go test`)
- Include relevant tests for changed behavior
- Keep the project's provenance and source-link requirements intact
- Be reviewed and approved by the repository owner before merge unless the repository owner is the actor making the change

The repository owner is exempt from the usual approval requirement and may merge directly to `main` for repository maintenance. All other contributors must use pull requests, required CI checks, and repo-owner review before merge.

Branch protection should require pull requests and status checks in GitHub settings, while allowing the repository owner to bypass them for direct maintenance work.

## Design proposals

Design proposals should include:

- Problem and motivation
- Constraints and assumptions
- Proposed approach
- Risks and tradeoffs
- Impact on retrieval quality, attribution, or user experience
- Testing or evaluation plan

## Evaluation and testing results

Evaluation runs are important to this project. If you run a prompt comparison or retrieval check, record it under `evaluations/` with the date and the exact prompt used.

We also accept structured community submissions under `community/`. Use a GitHub username directory and a UTC timestamp directory, for example:

```text
community/ideas/your-github-username/2026-08-30T090000Z/idea.md
community/test-results/your-github-username/2026-08-30T090000Z/result.md
community/design-proposals/your-github-username/2026-08-30T090000Z/proposal.md
```

The date must be in ISO 8601 UTC form with a `Z`, and the filename should remain a Markdown file. CI validates the directory naming convention with `scripts/check_contrib_paths.py`.

All claims about poem text, author attribution, or Ganjoor URLs must be supported by evidence. If the result is unverified, say so explicitly.

## Pull requests

PRs should include:

- Short summary of the change
- Why the change is needed
- Relevant files affected
- Test commands run
- Any evaluation notes or dataset additions
- Documentation updates if needed

Do not merge your own PR until it has been reviewed and approved.

## Project governance and maintainership

The repository owner is the initial maintainer. As the project grows, governance can be expanded through a documented maintainer policy and custom review requirements.

The repository owner is exempt from the usual approval requirement and may merge directly to `main` for repository maintenance. All other contributors must use pull requests, required CI checks, and repo-owner review before merge.

## Security

If you discover a security vulnerability, do not open a public issue. Follow the reporting process in [`SECURITY.md`](SECURITY.md).

## Code of conduct

The project expects respectful, constructive participation. See [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md).

## License

This project is licensed under the MIT License. See `LICENSE` for details.
