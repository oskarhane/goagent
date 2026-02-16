# Kubernetes Monitoring Agent Example

This example demonstrates how to build an AI agent that monitors Kubernetes cluster health and reports issues.

## What It Does

The agent:
- Queries pod status across all namespaces
- Checks service health
- Identifies deployments with unavailable replicas
- Monitors node status
- Provides a comprehensive health report with recommendations

## Prerequisites

1. **Kubernetes cluster access** - The agent needs a valid kubeconfig
2. **OpenAI API key** - Set in environment variable
3. **Go 1.25+**

## Quick Start

### 1. Set up environment

Create `.env` file:

```bash
OPENAI_API_KEY=sk-your-key-here
KUBECONFIG=/path/to/kubeconfig  # Optional, defaults to ~/.kube/config
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Run the agent

```bash
go run main.go
```

For debug output:

```bash
DEBUG=true go run main.go
```

## Example Output

```
=== Kubernetes Cluster Health Check ===
Started at: 2026-02-16T14:30:00Z

=== Monitoring Report ===
Cluster Health Status: HEALTHY

I've completed a comprehensive health check of your Kubernetes cluster. Here's what I found:

Pods (All Namespaces):
- Total pods checked: 45
- Running: 43
- Issues found: 2
  * nginx-deployment-abc123 in namespace prod: CrashLoopBackOff
  * worker-job-xyz789 in namespace default: ImagePullBackOff

Deployments:
- All deployments are at desired replica count
- No unavailable replicas detected

Nodes:
- All 5 nodes are in Ready state
- No node issues detected

Recommendations:
1. Investigate nginx-deployment-abc123 logs in prod namespace - appears to be application error
2. Check image pull secrets for worker-job-xyz789 - image may not exist or credentials invalid

Overall: Cluster is in good health with 2 minor pod issues requiring attention.

=== Execution Stats ===
Iterations: 8
Total tokens: 3421
Duration: 12.3s
Completed at: 2026-02-16T14:30:12Z
```

## Kubernetes Permissions

The agent needs read access to cluster resources. Use the provided RBAC manifest:

```bash
kubectl apply -f ../../deployments/k8s/base/rbac.yaml
```

This creates:
- ServiceAccount: `goagent`
- Role: Read access to pods, services, deployments, nodes
- RoleBinding: Binds role to service account

## Deploy as CronJob

To run this monitoring agent on a schedule:

```bash
# Update the manifest with your API key
kubectl create secret generic goagent-secrets \
  --from-literal=openai_api_key=sk-your-key

# Deploy as CronJob (runs every 15 minutes)
kubectl apply -f ../../deployments/k8s/examples/monitoring-agent.yaml
```

## Customization

### Change monitoring scope

Edit the `systemPrompt` in `main.go` to focus on specific resources:

```go
systemPrompt := `Check only pods in the production namespace and report any that are not Running.`
```

### Adjust monitoring frequency

For CronJob deployment, edit the schedule in `monitoring-agent.yaml`:

```yaml
spec:
  schedule: "*/15 * * * *"  # Every 15 minutes
```

### Add alerting

Integrate with alerting systems by parsing the agent's output:

```go
// After agent.Run()
if strings.Contains(result.Messages[0].Content, "CRITICAL") {
    sendAlert(result.Messages[0].Content)
}
```

## Troubleshooting

**"failed to create client: unable to load kubeconfig"**
- Set `KUBECONFIG` environment variable
- Or place kubeconfig at `~/.kube/config`

**"failed to list pods: forbidden"**
- Apply RBAC manifest: `kubectl apply -f ../../deployments/k8s/base/rbac.yaml`
- Ensure you're using the `goagent` service account

**"context deadline exceeded"**
- Increase timeout in K8s tool config
- Check cluster connectivity

## Related Examples

- [K8s Tool](../k8s-tool/) - Basic K8s tool usage
- [Incident Response](../incident-response/) - Investigation agent
- [Agent Basics](../agent-basic/) - Core agent concepts
