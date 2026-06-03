# quiccpy-start

A generic Python project template using `uv`, `structlog`, `pydantic`, and `pydantic-settings`.

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

## Configuration

Config is loaded from `config.yaml`. See `config.example.yaml` for all available fields.

### Env var overrides

Any config value can be overridden with an environment variable using the `APP_` prefix and `__` as the nested delimiter. 
Env vars take priority over `config.yaml`.

```bash
# Override logging level
APP_LOGGING__LEVEL=DEBUG
```

## Testing

```bash
make test
# or
uv run --group dev pytest
```

Tests live in `tests/` mirroring the `src/` structure.

## Linting

```bash
make lint
# or
uv run ruff check . --fix
```

## Docker

```bash
cd deployment
docker compose up
```

Volumes mount `logs/` and `config.yaml` from the project root into the container.
