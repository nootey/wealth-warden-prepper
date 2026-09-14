# Wealth Warden Prepper - Claude instructions

## Project context

- `prepper` is a Go CLI that parses bank statements (CSV and PDF) into a normalized JSON payload for import into Wealth Warden
- Parsing is split into a shared, best-effort engine (`pkg/statement`) and small per-bank rule sets (`pkg/rules/<bank>`) that only supply what the engine can't guess: header aliases and defaults for CSV, a line-format regex for PDF
- Supported banks: `nlb`, `n26`, `revolut`. Unrecognized banks fall back to `generic`, which runs the shared engine with no bank-specific hints
- To onboard a new bank: try `-bank generic` on its CSV first, see what it can't map, then add a rule set next to the existing ones. PDF layouts don't generalize, so those stay bank-specific from the start


## Workflow

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- Wait for explicit approval before writing any code or changing files
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## Development Guidelines

- For exploration tasks (finding files, grepping), prefer spawning Explore subagents rather than reading into main context
- DO NOT suggest service to service injections, unless absolutely necessary - present your reasoning if so
- Match existing code patterns and conventions even if you'd do it differently
- Build feature by feature, and write tests after each implementation, if applicable
  - Tests should be high impact only, do not cover everything
- Minimum code that solves the problem. Nothing speculative.
- When writing tests for different bank statements, anonymize data

## General guidelines
- Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify
- Don't assume. Don't hide confusion. Surface tradeoffs
- Define success criteria. Loop until verified
- Transform tasks into verifiable goals:
  - "Add validation" → "Write tests for invalid inputs, then make them pass"
  - "Fix the bug" → "Write a test that reproduces it, then make it pass"
  - "Refactor X" → "Ensure tests pass before and after"