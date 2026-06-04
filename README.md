# quiccpy-start

Helper script to migrate bank statements to a format, that Wealth Warden expects. Currently only supports some Slovene banks.

## Docs

Check out the docs for instructions on how to fetch statements from different banks.
Currently supported:
- [NLB](./docs/bank_statements/nlb.md)

## Requirements

- Python 3.11+
- [uv](https://docs.astral.sh/uv/)

## Setup

```bash
cp config.example.yaml config.yaml
uv sync --group dev
uv run pre-commit install
```

## Running

```bash
make run
# or
uv run python -m main
```

