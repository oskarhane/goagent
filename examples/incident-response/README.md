# Incident Response Agent Example

This example demonstrates how to build an AI agent that investigates and diagnoses production incidents by correlating data from multiple sources.

## What It Does

The agent:
- Takes an incident description as input
- Systematically investigates using multiple tools (K8s, HTTP, shell)
- Correlates data from different sources
- Identifies root cause
- Provides actionable recommendations

## Tools Available

1. **Kubernetes queries** - Check pod status, logs, deployments
2. **HTTP requests** - Query monitoring APIs, health endpoints
3. **Shell commands** - Run safe diagnostic commands (read-only)

## Prerequisites

1. **OpenAI API key** - Set in environment variable
2. **Kubernetes cluster access** - Optional, for K8s investigation
3. **Go 1.26+**

## Quick Start

### 1. Set up environment

Create `.env` file:

```bash
OPENAI_API_KEY=sk-your-key-here
KUBECONFIG=/path/to/kubeconfig  # Optional
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Run the agent

With default incident:

```bash
go run main.go
```

With custom incident description:

```bash
go run main.go "Database connections are timing out in production"
```

For debug output:

```bash
DEBUG=true go run main.go "API latency increased by 300%"
```

## Example Output

```
=== Incident Response Investigation ===
Started at: 2026-02-16T15:00:00Z
Incident: Users are reporting 500 errors from the production API. The errors started 10 minutes ago.

=== Investigation Report ===
I've completed the investigation. Here are my findings:

ROOT CAUSE IDENTIFIED:
The production API pods in namespace 'prod' are experiencing OOMKilled errors. 
3 out of 5 pods have been restarting repeatedly in the last 10 minutes.

EVIDENCE:
1. Kubernetes Query Results:
   - Pod 'api-deployment-7d8f9c-abc12' status: CrashLoopBackOff
   - Last restart: 2 minutes ago
   - Exit code: 137 (OOMKilled)
   - Memory limit: 512Mi, current usage before crash: 498Mi

2. HTTP Check Results:
   - Health endpoint (http://api.prod.svc/health) returning 503
   - Only 2/5 pods responding
   - Average response time: 8.5s (normally <100ms)

3. System Metrics:
   - Memory pressure detected on nodes
   - Swap usage at 85%

ROOT CAUSE:
Memory leak or increased memory usage causing pods to exceed 512Mi limit,
leading to OOMKill events and cascading failures.

RECOMMENDATIONS:
1. IMMEDIATE: Increase pod memory limit from 512Mi to 1Gi
2. IMMEDIATE: Scale deployment from 5 to 8 replicas to handle load during recovery
3. SHORT-TERM: Investigate memory leak in application code
4. SHORT-TERM: Add memory profiling and alerting at 80% threshold
5. LONG-TERM: Implement circuit breakers to prevent cascade failures

Commands to execute:
```bash
kubectl -n prod set resources deployment api-deployment --limits=memory=1Gi
kubectl -n prod scale deployment api-deployment --replicas=8
```

Expected recovery time: 2-3 minutes after applying fixes.

=== Tool Usage Summary ===
Tools executed: 7

=== Execution Stats ===
Iterations: 12
Total tokens: 5234
Duration: 28.7s
Completed at: 2026-02-16T15:00:29Z
```

## Example Scenarios

### Database Connection Issues

```bash
go run main.go "Database connections timing out, application logs show 'connection pool exhausted'"
```

Agent will:
- Check pod logs for connection errors
- Query database service endpoints
- Check network policies
- Investigate connection pool configuration

### High Latency

```bash
go run main.go "API response time increased from 100ms to 3s"
```

Agent will:
- Check pod resource usage (CPU/memory)
- Query external dependencies
- Check for recent deployments
- Analyze traffic patterns

### Service Unavailable

```bash
go run main.go "Service returning 503 errors intermittently"
```

Agent will:
- Check pod readiness and health
- Verify service endpoints
- Check load balancer configuration
- Investigate network issues

## Customization

### Add Custom Monitoring APIs

Edit `main.go` to add your monitoring endpoints:

```go
// Example: Query Prometheus
task := `
Investigate this incident: %s

First, query our Prometheus API at http://prometheus.monitoring.svc:9090/api/v1/query 
to get current error rates and latency metrics.
`
```

### Adjust Tool Permissions

Modify allowed shell commands for your environment:

```go
shellHandler := shell.NewHandler(&shell.Config{
    AllowedCommands: []string{
        "kubectl", "docker", "systemctl", // Add your tools
    },
})
```

### Multi-Turn Investigation

For interactive investigation, use conversation history:

```go
// First query
result1 := a.Run(ctx, "Investigate 500 errors", nil)

// Follow-up with context
result2 := a.Run(ctx, "Can you check the database connection pool?", &agent.RunOptions{
    History: result1.Messages,
})
```

## Kubernetes Permissions

The agent needs read access to cluster resources. Apply RBAC:

```bash
kubectl apply -f ../../deployments/k8s/base/rbac.yaml
```

For log access, add to Role:

```yaml
- apiGroups: [""]
  resources: ["pods/log"]
  verbs: ["get", "list"]
```

## Safety Features

### Shell Command Safety

Only allowed commands can execute:
- Blocklist prevents destructive operations (`rm`, `dd`, `mkfs`)
- Allowlist restricts to diagnostic tools
- Output size limited to prevent memory exhaustion
- Timeout prevents hung processes

### HTTP Safety

- Request timeout: 30s
- Response size limit: 10MB
- No automatic redirects
- User-Agent header set

### Kubernetes Safety

- Read-only operations only
- RBAC permissions enforced by cluster
- Operation timeout: 30s

## Deploy as Service

Run as an HTTP service for on-demand investigations:

```bash
# Build image
docker build -t incident-agent .

# Deploy to Kubernetes
kubectl apply -f ../../deployments/k8s/examples/incident-response.yaml
```

Note: Service mode requires HTTP endpoint implementation (future enhancement).

## Troubleshooting

**"OPENAI_API_KEY environment variable is required"**
- Set API key in `.env` or export to environment

**"Failed to create K8s client"**
- Set KUBECONFIG or place config at `~/.kube/config`
- Verify cluster connectivity

**"command not found: kubectl"**
- Shell commands run in container/local environment
- Add kubectl to PATH or allowlist

**Agent stops after max iterations**
- Increase `MaxIterations` in config
- Current limit: 20 iterations

## Related Examples

- [K8s Monitoring](../k8s-monitoring/) - Proactive monitoring
- [K8s Tool](../k8s-tool/) - Basic K8s queries
- [HTTP Tool](../http-tool/) - HTTP API calls
- [Shell Tool](../shell-tool/) - Command execution
