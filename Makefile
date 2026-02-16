# GoAgent SDK Makefile

.PHONY: help build test lint clean install docs dev-setup

# Default target
help: ## Show this help message
	@echo "GoAgent SDK - Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development setup
dev-setup: ## Install development dependencies
	@echo "Installing development dependencies..."
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/godoc@latest
	@echo "Development dependencies installed!"

# Build
build: ## Build the CLI binary
	@echo "Building goagent..."
	@go build -o bin/goagent ./cmd/goagent
	@echo "Built: bin/goagent"

install: ## Install the CLI binary to $GOPATH/bin
	@echo "Installing goagent..."
	@go install ./cmd/goagent
	@echo "Installed goagent to $(shell go env GOPATH)/bin"

# Testing
test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Linting
lint: ## Run linting
	@echo "Running linter..."
	@PATH=$(shell go env GOPATH)/bin:$(PATH) golangci-lint run

lint-fix: ## Run linting with auto-fix
	@echo "Running linter with auto-fix..."
	@PATH=$(shell go env GOPATH)/bin:$(PATH) golangci-lint run --fix

# Code formatting
fmt: ## Format code
	@echo "Formatting code..."
	@PATH=$(shell go env GOPATH)/bin:$(PATH) goimports -w . || go fmt ./...
	@go fmt ./...

# Documentation
docs: ## Generate documentation
	@echo "Starting documentation server..."
	@echo "Visit http://localhost:6060/pkg/github.com/oskarhane/goagent/"
	@godoc -http=:6060

# Clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# CI targets
ci-test: ## Run tests for CI
	@go test -v -race -coverprofile=coverage.out ./...

ci-lint: ## Run linting for CI  
	@PATH=$(shell go env GOPATH)/bin:$(PATH) golangci-lint run --timeout=5m

# Docker
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t goagent:latest --build-arg VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev") .
	@echo "Docker image built: goagent:latest"

docker-build-service: ## Build service variant Docker image
	@echo "Building service variant Docker image..."
	@docker build -t goagent:service -f deployments/docker/Dockerfile.agent-service --build-arg VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev") .
	@echo "Docker image built: goagent:service"

docker-build-cronjob: ## Build cronjob variant Docker image
	@echo "Building cronjob variant Docker image..."
	@docker build -t goagent:cronjob -f deployments/docker/Dockerfile.agent-cronjob --build-arg VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev") .
	@echo "Docker image built: goagent:cronjob"

docker-build-all: docker-build docker-build-service docker-build-cronjob ## Build all Docker image variants
	@echo "All Docker images built!"

docker-run: ## Run Docker container locally
	@echo "Running Docker container..."
	@docker run --rm -it --env-file .env goagent:latest

docker-compose-up: ## Start services with Docker Compose
	@echo "Starting Docker Compose services..."
	@cd deployments/docker && docker-compose up -d
	@echo "Services started! Run 'make docker-compose-logs' to view logs"

docker-compose-down: ## Stop Docker Compose services
	@echo "Stopping Docker Compose services..."
	@cd deployments/docker && docker-compose down
	@echo "Services stopped!"

docker-compose-logs: ## View Docker Compose logs
	@cd deployments/docker && docker-compose logs -f

# Kubernetes
k8s-validate: ## Validate Kubernetes manifests
	@echo "Validating Kubernetes manifests..."
	@python3 -c "import yaml; [list(yaml.safe_load_all(open(f))) for f in ['deployments/k8s/base/configmap.yaml', 'deployments/k8s/base/cronjob.yaml', 'deployments/k8s/base/deployment.yaml', 'deployments/k8s/base/rbac.yaml', 'deployments/k8s/base/secret.yaml']]" && echo "✓ All manifests valid"

k8s-apply-base: ## Apply base Kubernetes manifests
	@echo "Applying base Kubernetes manifests..."
	@kubectl apply -f deployments/k8s/base/rbac.yaml
	@kubectl apply -f deployments/k8s/base/configmap.yaml
	@kubectl apply -f deployments/k8s/base/secret.yaml
	@echo "Base manifests applied!"

k8s-apply-deployment: ## Apply Deployment (service mode)
	@echo "Applying Deployment..."
	@kubectl apply -f deployments/k8s/base/deployment.yaml
	@echo "Deployment applied!"

k8s-apply-cronjob: ## Apply CronJob (scheduled mode)
	@echo "Applying CronJob..."
	@kubectl apply -f deployments/k8s/base/cronjob.yaml
	@echo "CronJob applied!"

k8s-apply-monitoring: ## Apply monitoring manifests
	@echo "Applying monitoring manifests..."
	@kubectl apply -f deployments/k8s/monitoring/
	@echo "Monitoring manifests applied!"

k8s-delete: ## Delete all Kubernetes resources
	@echo "Deleting Kubernetes resources..."
	@kubectl delete -f deployments/k8s/base/ --ignore-not-found=true
	@echo "Resources deleted!"

k8s-logs: ## View logs from goagent pods
	@kubectl logs -l app=goagent -f

k8s-status: ## Check status of goagent resources
	@echo "GoAgent Resources:"
	@kubectl get all -l app=goagent
	@echo ""
	@echo "Secrets:"
	@kubectl get secret goagent-secrets
	@echo ""
	@echo "ConfigMaps:"
	@kubectl get configmap goagent-config

# Development workflow
dev: fmt lint test build ## Run development workflow (fmt, lint, test, build)
	@echo "Development workflow complete!"