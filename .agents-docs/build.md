# Build System

BUILD SYSTEMS: [Go modules (go.mod/go.sum), Make, Docker, GoReleaser (GitHub Actions)]
BUILD COMMANDS: [`make build`, `go build -o bin/goagent ./cmd/goagent`, `make docker-build`, `make docker-build-service`, `make docker-build-cronjob`, `make docker-build-all`, `goreleaser release --clean`]
LINT COMMANDS: [`make lint`, `make ci-lint`, `golangci-lint run --timeout=5m`]
FORMAT COMMANDS: [`make fmt`, `goimports -w .`, `go fmt ./...`]
BUNDLING: [Docker images via Dockerfile + Buildx, GoReleaser binaries]

---

*This file is part of the AGENTS.md documentation system.*
