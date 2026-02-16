# Docker Deployment Example

This example demonstrates how to deploy a GoAgent application using Docker.

## Prerequisites

- Docker installed and running
- OpenAI API key or Google Cloud credentials
- Basic understanding of Docker concepts

## Quick Start

### 1. Build the Image

From the repository root:

```bash
# Using make (recommended)
make docker-build

# Or using docker directly
docker build -t goagent:latest .
```

### 2. Set Up Environment

Create a `.env` file with your configuration:

```bash
cp deployments/docker/.env.example .env
# Edit .env with your API keys
```

### 3. Run the Container

```bash
# Using make
make docker-run

# Or using docker directly
docker run --rm --env-file .env goagent:latest --version
```

## Example: Kubernetes Monitoring Agent

This example shows a containerized agent that monitors Kubernetes clusters.

### Setup

1. **Prepare your kubeconfig**

```bash
# Ensure your kubeconfig has correct permissions
chmod 600 ~/.kube/config
```

2. **Create environment file** (`.env`):

```bash
OPENAI_API_KEY=sk-your-key-here
LOG_LEVEL=info
MAX_ITERATIONS=10
```

3. **Run the agent with kubeconfig access**:

```bash
docker run --rm \
  --env-file .env \
  -v ~/.kube/config:/home/goagent/.kube/config:ro \
  goagent:latest
```

## Example: Scheduled Monitoring with CronJob Image

Use the cronjob variant for periodic execution:

### 1. Build CronJob Image

```bash
make docker-build-cronjob
```

### 2. Run Periodically

```bash
# Run every 5 minutes using cron
echo "*/5 * * * * docker run --rm --env-file /path/to/.env goagent:cronjob run" | crontab -
```

Or use Docker's restart policies:

```bash
docker run -d \
  --name goagent-monitor \
  --restart unless-stopped \
  --env-file .env \
  -v ~/.kube/config:/home/goagent/.kube/config:ro \
  goagent:cronjob
```

## Example: Long-Running Service

Use the service variant for continuous operation:

### 1. Build Service Image

```bash
make docker-build-service
```

### 2. Run as Service

```bash
docker run -d \
  --name goagent-service \
  --restart unless-stopped \
  -p 8080:8080 \
  --env-file .env \
  goagent:service
```

### 3. Check Health

```bash
# Check container health
docker inspect --format='{{.State.Health.Status}}' goagent-service

# View logs
docker logs -f goagent-service
```

## Docker Compose Example

For local development with multiple configurations:

### 1. Start Services

```bash
cd deployments/docker
docker-compose up -d
```

### 2. View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f agent-openai
```

### 3. Stop Services

```bash
docker-compose down
```

## Advanced Usage

### Multi-Platform Build

Build for both AMD64 and ARM64:

```bash
# Enable buildx
docker buildx create --use

# Build for multiple platforms
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t goagent:multi \
  --load \
  .
```

### Custom Configuration

Override environment variables:

```bash
docker run --rm \
  -e OPENAI_API_KEY=your-key \
  -e LOG_LEVEL=debug \
  -e MAX_ITERATIONS=5 \
  goagent:latest
```

### Volume Mounts

Mount additional configuration:

```bash
docker run --rm \
  --env-file .env \
  -v $(pwd)/custom-config:/app/config:ro \
  -v $(pwd)/logs:/app/logs \
  goagent:latest
```

## Resource Management

### Set Resource Limits

```bash
docker run --rm \
  --memory=512m \
  --cpus=1 \
  --env-file .env \
  goagent:latest
```

### Monitor Resource Usage

```bash
docker stats goagent-service
```

## Troubleshooting

### Issue: Container Exits Immediately

**Solution**: Check logs for errors
```bash
docker logs container-name
```

### Issue: Permission Denied on Mounted Files

**Solution**: Fix file permissions
```bash
chmod 644 ~/.kube/config
chown 1000:1000 ~/.kube/config
```

### Issue: Health Check Failing

**Solution**: Verify the binary works
```bash
docker exec container-name /app/goagent --version
```

### Issue: Network Connectivity

**Solution**: Check DNS and network settings
```bash
docker run --rm --env-file .env goagent:latest sh -c "ping -c 1 api.openai.com"
```

## Security Best Practices

1. **Never commit `.env` files** with secrets
2. **Use Docker secrets** in production:
   ```bash
   echo "sk-your-key" | docker secret create openai_key -
   ```
3. **Run as non-root** (already configured)
4. **Use read-only filesystem** when possible:
   ```bash
   docker run --read-only --env-file .env goagent:latest
   ```
5. **Scan images** for vulnerabilities:
   ```bash
   docker scan goagent:latest
   ```

## Next Steps

- Deploy to Kubernetes: See `../k8s-deployment/`
- Set up CI/CD: See `.github/workflows/docker.yml`
- Configure monitoring: See `../monitoring/`
