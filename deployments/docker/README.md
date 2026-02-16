# Docker Deployment Guide

This directory contains Docker configurations for deploying GoAgent in various patterns.

## Quick Start

### Build the Image

```bash
# Build with default version
docker build -t goagent:latest .

# Build with specific version
docker build -t goagent:v1.0.0 --build-arg VERSION=v1.0.0 .
```

### Run a Container

```bash
# Basic run with OpenAI
docker run --rm \
  -e OPENAI_API_KEY=your-key-here \
  goagent:latest

# Run with environment file
docker run --rm \
  --env-file .env \
  goagent:latest
```

## Available Dockerfiles

### 1. `Dockerfile` (Root)
**Purpose**: Base production-ready image  
**Use Case**: General-purpose agent deployment  
**Features**:
- Multi-stage build for minimal image size
- Non-root user (UID 1000)
- Health checks enabled
- Environment variable configuration

**Build:**
```bash
docker build -t goagent:latest -f Dockerfile .
```

### 2. `Dockerfile.agent-service`
**Purpose**: Long-running service mode  
**Use Case**: HTTP endpoints, message queue consumers  
**Features**:
- Optimized for continuous operation
- Includes curl for health checks
- Metrics and tracing enabled by default
- Port 8080 exposed

**Build:**
```bash
docker build -t goagent:service -f deployments/docker/Dockerfile.agent-service .
```

### 3. `Dockerfile.agent-cronjob`
**Purpose**: Scheduled/periodic execution  
**Use Case**: Kubernetes CronJob, scheduled tasks  
**Features**:
- Minimal dependencies for fast startup
- No health check (runs to completion)
- Extended timeout (10 minutes default)
- Optimized resource usage

**Build:**
```bash
docker build -t goagent:cronjob -f deployments/docker/Dockerfile.agent-cronjob .
```

## Docker Compose

Use `docker-compose.yml` for local development and testing.

### Start Services

```bash
# Start all services
cd deployments/docker
docker-compose up -d

# Start specific service
docker-compose up -d agent-openai

# View logs
docker-compose logs -f agent-openai
```

### Configuration

Create a `.env` file in the `deployments/docker` directory:

```bash
# OpenAI Configuration
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-5.1

# Vertex AI Configuration
GOOGLE_CLOUD_PROJECT=your-project-id
GOOGLE_CREDENTIALS_PATH=./credentials/key.json
VERTEX_LOCATION=us-central1
VERTEX_MODEL=gemini-2.5-pro

# Agent Configuration
LOG_LEVEL=info
MAX_ITERATIONS=10
TIMEOUT=300

# Observability
ENABLE_TRACING=false
OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4318
```

## Environment Variables

### Provider Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `OPENAI_API_KEY` | OpenAI API key | - | For OpenAI |
| `OPENAI_MODEL` | OpenAI model name | `gpt-5.1` | No |
| `GOOGLE_CLOUD_PROJECT` | GCP project ID | - | For Vertex AI |
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to service account key | - | For Vertex AI |
| `VERTEX_LOCATION` | Vertex AI location | `us-central1` | No |
| `VERTEX_MODEL` | Vertex AI model name | `gemini-2.5-pro` | No |

### Agent Configuration

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` | No |
| `MAX_ITERATIONS` | Maximum agent iterations | `10` | No |
| `TIMEOUT` | Timeout in seconds | `300` | No |
| `ENABLE_METRICS` | Enable Prometheus metrics | `false` | No |
| `ENABLE_TRACING` | Enable OpenTelemetry tracing | `false` | No |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | OTLP endpoint URL | - | If tracing enabled |

## Best Practices

### Security

1. **Never commit secrets**: Use `.env` files or secret management
2. **Run as non-root**: All images use UID 1000 by default
3. **Minimal base image**: Alpine Linux for smaller attack surface
4. **Read-only filesystem**: Consider using `--read-only` flag

```bash
docker run --rm --read-only \
  -e OPENAI_API_KEY=your-key \
  goagent:latest
```

### Resource Limits

Always set resource limits in production:

```bash
docker run --rm \
  --memory=512m \
  --cpus=1 \
  -e OPENAI_API_KEY=your-key \
  goagent:latest
```

### Health Checks

Monitor container health:

```bash
# Check health status
docker inspect --format='{{.State.Health.Status}}' container-name

# View health check logs
docker inspect --format='{{range .State.Health.Log}}{{.Output}}{{end}}' container-name
```

### Logging

Structured JSON logs are written to stdout:

```bash
# Follow logs
docker logs -f container-name

# Filter by level
docker logs container-name 2>&1 | grep '"level":"error"'
```

## Advanced Usage

### Multi-Platform Builds

Build for multiple architectures:

```bash
# Enable buildx
docker buildx create --use

# Build for amd64 and arm64
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t goagent:latest \
  --push \
  .
```

### Custom Entrypoint

Override the default entrypoint for custom workflows:

```bash
docker run --rm \
  --entrypoint /bin/sh \
  goagent:latest \
  -c "echo 'Custom command'"
```

### Development Mode

Mount source code for development:

```bash
docker run --rm \
  -v $(pwd):/app/src:ro \
  -e OPENAI_API_KEY=your-key \
  goagent:dev
```

## Troubleshooting

### Container Exits Immediately

Check logs for errors:
```bash
docker logs container-name
```

### Permission Denied

Ensure mounted volumes have correct permissions:
```bash
chmod 644 ~/.kube/config
```

### Health Check Failing

Verify the binary works:
```bash
docker exec container-name /app/goagent --version
```

### Out of Memory

Increase memory limit:
```bash
docker run --memory=1g goagent:latest
```

## Next Steps

- Deploy to Kubernetes: See `../k8s/README.md`
- Configure observability: See `../monitoring/README.md`
- Set up CI/CD: See `.github/workflows/docker.yml`
