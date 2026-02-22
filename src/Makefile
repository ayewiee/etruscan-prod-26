# Run from src/ (module root). Requires: Go, Docker (for e2e).

.PHONY: test-unit test-unit-verbose test-e2e test-e2e-verbose test-e2e-ui test coverage coverage-html deps ci-unit ci-e2e cobertura

# Unit tests only (no DB, no containers). Use in CI for fast feedback.
test-unit:
	go test -short -count=1 ./...

# Unit tests with verbose output (each package and test).
test-unit-verbose:
	go test -short -count=1 -v ./...

# E2E tests (Testcontainers: Postgres + Redis). Requires Docker.
test-e2e:
	go test -count=1 -timeout 10m -v ./tests/e2e/...

# Alias for clarity (e2e is already verbose).
test-e2e-verbose: test-e2e

# E2E with gotestsum UI (install: go install gotest.tools/gotestsum@latest).
test-e2e-ui:
	gotestsum --format testdox -- -count=1 -timeout 10m ./tests/e2e/...

# All tests (unit + e2e)
test: test-unit test-e2e

# Coverage from unit tests only (excludes tests/e2e so no Docker needed).
# Output: coverage.out, and optionally coverage.html.
coverage:
	go test -short -count=1 -coverprofile=coverage.out -covermode=atomic $$(go list ./... | grep -v '/tests/e2e$$')
	@echo "Coverage written to coverage.out"
	@go tool cover -func=coverage.out | tail -1

# Human-readable HTML report. Run after 'make coverage'.
coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "Open coverage.html in a browser"

# Optional: coverage including e2e (slower, broader).
coverage-all:
	go test -count=1 -timeout 10m -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | tail -1

# --- CI targets (used by GitLab CI; require gotestsum; e2e also uses gocover-cobertura) ---
# Unit tests + JUnit. Writes junit-unit.xml (no coverage in CI; coverage comes from e2e).
ci-unit:
	@which gotestsum >/dev/null || (echo "install gotestsum: go install gotest.tools/gotestsum@latest" && exit 1)
	gotestsum --junitfile junit-unit.xml -- -short -count=1 $$(go list ./... | grep -v '/tests/e2e$$')

# E2E tests + JUnit + coverage. -coverpkg=./... includes app code (handlers, usecases) in profile.
ci-e2e:
	@which gotestsum >/dev/null || (echo "install gotestsum: go install gotest.tools/gotestsum@latest" && exit 1)
	gotestsum --junitfile junit-e2e.xml -- -count=1 -timeout 10m -coverprofile=coverage.out -covermode=atomic -coverpkg=./... ./tests/e2e/...
	@go tool cover -func=coverage.out | tail -1

# Cobertura XML from coverage.out (for GitLab coverage MR widget). Run after coverage or ci-e2e.
# Install: go install github.com/boumenot/gocover-cobertura@latest
cobertura:
	@test -f coverage.out || (echo "run 'make coverage' or 'make ci-e2e' first" && exit 1)
	@which gocover-cobertura >/dev/null || (echo "install gocover-cobertura: go install github.com/boumenot/gocover-cobertura@latest" && exit 1)
	gocover-cobertura < coverage.out > cobertura.xml
	@echo "cobertura.xml written"

# Fetch dependencies (e.g. after adding e2e deps).
deps:
	go mod tidy
	go mod download
