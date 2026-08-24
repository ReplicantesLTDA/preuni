.PHONY: help up down logs fmt lint type-check test-unit test-int test-golden-fake test-golden-real clean install correct correct-sample correct-fake

# Auto-load .env if present (Make exports each var to child processes).
ifneq (,$(wildcard ./.env))
include .env
export
endif

# Override on CLI: make correct-sample ESSAY=path THEME_FILE=path
ESSAY ?= tests/golden/essay_br/test/0001/essay.txt
THEME_FILE ?= tests/golden/essay_br/test/0001/prompt_theme.txt

# Display help message
help:
	@echo "Corretor Redação - Automated ENEM Essay Correction API"
	@echo ""
	@echo "Available commands:"
	@echo "  make up                  - Start development environment (docker-compose up)"
	@echo "  make down                - Stop development environment"
	@echo "  make logs                - View logs from running containers"
	@echo "  make install             - Install dependencies using uv"
	@echo "  make fmt                 - Format code with ruff"
	@echo "  make lint                - Run linters (ruff, mypy, import-linter)"
	@echo "  make type-check          - Run type checking with mypy"
	@echo "  make test-unit           - Run unit tests"
	@echo "  make test-int            - Run integration tests"
	@echo "  make test-golden-fake    - Run golden dataset tests with fake provider"
	@echo "  make test-golden-real    - Run golden dataset tests with real Ollama provider"
	@echo "  make correct-sample      - Correct the bundled sample essay via real Ollama (.env loads OLLAMA_*)"
	@echo "  make correct ESSAY=path THEME_FILE=path  - Correct any essay via real Ollama"
	@echo "  make correct-fake        - Same as correct, but bypasses the LLM (plumbing smoke)"
	@echo "  make clean               - Remove __pycache__, build artifacts, and test reports"

# Install dependencies with uv
install:
	@echo "Installing dependencies..."
	uv pip install -e ".[dev]"
	@echo "✓ Dependencies installed"

# Start development environment
up:
	@echo "Starting development environment..."
	docker-compose -f deploy/docker-compose.yml up -d
	@echo "✓ Environment started"
	@echo "  API: http://localhost:8000"
	@echo "  Docs: http://localhost:8000/docs"
	@echo "  Postgres: localhost:5432"

# Stop development environment
down:
	@echo "Stopping development environment..."
	docker-compose -f deploy/docker-compose.yml down
	@echo "✓ Environment stopped"

# View logs
logs:
	docker-compose -f deploy/docker-compose.yml logs -f

# Format code
fmt:
	@echo "Formatting code..."
	ruff format src tests
	@echo "✓ Code formatted"

# Run all linters
lint: fmt type-check import-lint
	@echo "✓ All linters passed"

# Type checking with mypy
type-check:
	@echo "Running type checking..."
	mypy src tests
	@echo "✓ Type checking passed"

# Import linter
import-lint:
	@echo "Running import linter..."
	lint-imports --root src
	@echo "✓ Import linter passed"

# Unit tests
test-unit:
	@echo "Running unit tests..."
	pytest tests/unit -v --cov=src --cov-report=html --cov-report=term-missing
	@echo "✓ Unit tests passed"

# Integration tests
test-int:
	@echo "Running integration tests..."
	pytest tests/integration -v -m integration
	@echo "✓ Integration tests passed"

# Golden dataset tests with fake provider (every PR, plumbing only)
test-golden-fake:
	@echo "Running golden dataset tests (fake provider)..."
	GOLDEN_PROVIDER=fake uv run pytest tests/golden -v -s
	@echo "✓ Golden dataset tests passed"

# Golden dataset tests with real provider (Article IX merge gate)
test-golden-real:
	@echo "Running golden dataset tests (real Ollama Cloud)..."
	GOLDEN_PROVIDER=auto uv run pytest tests/golden/test_golden_dataset.py -v -s
	@echo "✓ Golden dataset tests passed"

# Correct ONE essay via the walking-skeleton CLI against the real LLM provider.
# Usage: make correct ESSAY=/path/to/essay.txt THEME_FILE=/path/to/theme.txt
correct:
	@title=$$(head -1 "$(THEME_FILE)"); \
	context=$$(awk 'f{print} /^---$$/{f=1}' "$(THEME_FILE)"); \
	[ -z "$$context" ] && context=$$(tail -n +2 "$(THEME_FILE)"); \
	uv run python -m src.scripts.correct_essay \
	  --essay "$(ESSAY)" \
	  --theme-title "$$title" \
	  --theme-context "$$context" \
	  --provider ollama

# Same as `correct` but with the bundled walking-skeleton sample.
correct-sample: correct

# Pipeline smoke without LLM (returns 1000/1000 canned output)
correct-fake:
	@title=$$(head -1 "$(THEME_FILE)"); \
	context=$$(awk 'f{print} /^---$$/{f=1}' "$(THEME_FILE)"); \
	[ -z "$$context" ] && context=$$(tail -n +2 "$(THEME_FILE)"); \
	uv run python -m src.scripts.correct_essay \
	  --essay "$(ESSAY)" \
	  --theme-title "$$title" \
	  --theme-context "$$context" \
	  --provider fake-perfect

# Clean up
clean:
	@echo "Cleaning up..."
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete
	find . -type f -name "*.pyo" -delete
	rm -rf build dist *.egg-info
	rm -rf .pytest_cache .mypy_cache htmlcov .coverage
	@echo "✓ Cleanup complete"

# Development server (for local testing)
dev:
	@echo "Starting API development server..."
	uvicorn src.api.main:app --reload --host 0.0.0.0 --port 8000

# Database migrations
db-migrate:
	@echo "Running database migrations..."
	alembic upgrade head
	@echo "✓ Migrations applied"

# Database rollback
db-rollback:
	@echo "Rolling back last migration..."
	alembic downgrade -1
	@echo "✓ Migration rolled back"

# Create new migration
db-new-migration:
	@echo "Creating new migration..."
	@echo "Provide migration name: "; \
	read name; \
	alembic revision --autogenerate -m "$$name"
	@echo "✓ Migration created"

# Run all tests
test: test-unit test-int
	@echo "✓ All tests passed"

# Full CI pipeline locally
ci: clean install lint test-golden-fake
	@echo "✓ CI pipeline complete"
