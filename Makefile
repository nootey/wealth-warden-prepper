default: run

run:
	python -m main

lint:
	uv run ruff check . --fix

test:
	uv run --group dev pytest