# Wealth Warden Prepper - Claude instructions

## Project context

- `prepper` is a Go CLI that parses bank statements (CSV and PDF) into a normalized JSON payload for import into Wealth Warden
- Parsing is split into a shared, best-effort engine (`pkg/statement`) and small per-bank rule sets (`pkg/rules/<bank>`) that only supply what the engine can't guess: header aliases and defaults for CSV, a line-format regex for PDF
- Supports a sub-set of banks. Unrecognized banks fall back to `generic`, which runs the shared engine with no bank-specific hints
- To onboard a new bank: try `-bank generic` on its CSV first, see what it can't map, then add a rule set next to the existing ones. PDF layouts don't generalize, so those stay bank-specific from the start

## Rules

- DO NOT suggest service to service injections, unless absolutely necessary - present your reasoning if so
- Match existing repository code patterns and conventions. If you'd do it differently, suggest
- Minimize helpers in any domain/service files. If they are needed, create them in utils package.
- DO NOT create separate test files, use shared per domain/service ones.
