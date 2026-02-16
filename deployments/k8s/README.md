# Kubernetes Deployment Guide

This directory contains Kubernetes manifests for deploying GoAgent in various patterns.

## Quick Start

### Prerequisites

- Kubernetes cluster (1.24+)
- kubectl configured
- Docker image built and pushed to registry

### Deploy with Default Configuration

```bash
# Create namespace (optional)
kubectl create namespace goagent

# Deploy base resources
kubectl apply -f base/

# Verify deployment
kubectl get pods -l app=goagent
kubectl logs -l app=goagent -f
```

## Directory Structure

```
k8s/
├── base/              # Core deployment manifests
│   ├── rbac.yaml      # ServiceAccount, Role, RoleBinding
│   ├── configmap.yaml # Configuration values
│   ├── secret.yaml    # Secret template (do not commit real values)
│   ├── deployment.yaml # Long-running service deployment
│   └── cronjob.yaml   # Scheduled job deployment
├── examples/          # Example agent configurations
│   ├── monitoring-agent.yaml
│   └── incident-response.yaml
└── monitoring/        # Observability integration
    ├── servicemonitor.yaml
    ├── podmonitor.yaml
    └── otel-collector.yaml
```

## Deployment Patterns

### 1. Service Deployment (Long-Running)

For agents that run continuously (HTTP endpoints, queue consumers, etc.):

```bash
kubectl apply -f base/rbac.yaml
kubectl apply -f base/configmap.yaml
kubectl apply -f base/secret.yaml  # After configuring secrets
kubectl apply -f base/deployment.yaml
```

**Use Cases:**
- HTTP API endpoints for agent invocation
- Message queue consumers
- Webhook handlers
- Real-time monitoring agents

### 2. CronJob Deployment (Scheduled)

For agents that run on a schedule:

```bash
kubectl apply -f base/rbac.yaml
kubectl apply -f base/configmap.yaml
kubectl apply -f base/secret.yaml  # After configuring secrets
kubectl apply -f base/cronjob.yaml
```

**Use Cases:**
- Periodic health checks
- Scheduled reports
- Batch processing tasks
- Regular maintenance operations

### 3. Job Deployment (One-Time)

For one-time agent execution:

```bash
kubectl create job --from=cronjob/goagent-scheduled goagent-adhoc
kubectl logs -l job-name=goagent-adhoc -f
```

## Configuration

### Secrets

Create secrets before deploying:

```bash
# Option 1: From literals
kubectl create secret generic goagent-secrets \
  --from-literal=openai_api_key=sk-... \
  --from-literal=google_project_id=my-project

# Option 2: From files
kubectl create secret generic goagent-secrets \
  --from-file=google_credentials=./service-account.json \
  --from-literal=openai_api_key=sk-...

# Option 3: Using sealed-secrets (recommended for GitOps)
# See: https://github.com/bitnami-labs/sealed-secrets
```

### ConfigMap

Edit `base/configmap.yaml` to customize agent behavior:

```yaml
data:
  log_level: "info"          # debug, info, warn, error
  max_iterations: "10"       # Maximum agent reasoning loops
  timeout: "300"             # Timeout in seconds
  openai_model: "gpt-4-turbo-preview"
  vertex_model: "gemini-1.5-pro"
```

Apply changes:

```bash
kubectl apply -f base/configmap.yaml
kubectl rollout restart deployment/goagent  # If using deployment
```

### Environment Variables

Common environment variables (set via ConfigMap/Secret):

| Variable | Source | Description | Required |
|----------|--------|-------------|----------|
| `OPENAI_API_KEY` | Secret | OpenAI API key | For OpenAI |
| `GOOGLE_CLOUD_PROJECT` | Secret | GCP project ID | For Vertex AI |
| `GOOGLE_APPLICATION_CREDENTIALS` | Volume | Path to service account key | For Vertex AI |
| `LOG_LEVEL` | ConfigMap | Logging verbosity | No |
| `MAX_ITERATIONS` | ConfigMap | Max reasoning iterations | No |
| `TIMEOUT` | ConfigMap | Timeout in seconds | No |

## RBAC Configuration

### Namespace-Scoped Access

The default `base/rbac.yaml` provides read-only access to resources within a namespace:

```yaml
# Resources granted:
- pods, services, configmaps, endpoints, events, namespaces (core)
- deployments, replicasets, statefulsets, daemonsets (apps)
- jobs, cronjobs (batch)
```

### Cluster-Wide Access

For agents that need to query across all namespaces, uncomment the ClusterRole section in `base/rbac.yaml`:

```bash
# Edit base/rbac.yaml and uncomment ClusterRole/ClusterRoleBinding
kubectl apply -f base/rbac.yaml
```

**Security Warning:** Only grant cluster-wide access if absolutely necessary. Use namespace-scoped roles when possible.

### Custom RBAC

Create a custom role for specific resource access:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: goagent-custom
rules:
- apiGroups: [""]
  resources: ["pods", "pods/log"]  # Add pods/log for log access
  verbs: ["get", "list", "watch"]
- apiGroups: [""]
  resources: ["secrets"]           # Add secrets if needed
  verbs: ["get"]
```

## Resource Configuration

### Resource Limits

Default resource allocation:

**Deployment (Service):**
```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 1000m
    memory: 512Mi
```

**CronJob:**
```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 256Mi
```

Adjust based on your workload:

```bash
# Edit deployment/cronjob manifest
kubectl edit deployment goagent
# Or apply updated manifest
kubectl apply -f base/deployment.yaml
```

### Horizontal Pod Autoscaling

For service deployments with variable load:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: goagent-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: goagent
  minReplicas: 1
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

Apply HPA:

```bash
kubectl apply -f hpa.yaml
kubectl get hpa
```

## Observability

### Prometheus Metrics

If using Prometheus Operator, apply ServiceMonitor:

```bash
kubectl apply -f monitoring/servicemonitor.yaml
```

For CronJobs, use PodMonitor:

```bash
kubectl apply -f monitoring/podmonitor.yaml
```

Verify scraping:

```bash
kubectl get servicemonitor
kubectl get podmonitor
```

### OpenTelemetry Tracing

Deploy OpenTelemetry Collector:

```bash
kubectl apply -f monitoring/otel-collector.yaml
```

Enable tracing in agent:

```bash
# Edit deployment env
- name: ENABLE_TRACING
  value: "true"
- name: OTEL_EXPORTER_OTLP_ENDPOINT
  value: "http://otel-collector:4318"

kubectl apply -f base/deployment.yaml
```

View traces in your backend (Jaeger, Tempo, etc.)

### Logging

Logs are written to stdout in JSON format. View logs:

```bash
# Recent logs
kubectl logs -l app=goagent

# Follow logs
kubectl logs -l app=goagent -f

# Specific container
kubectl logs deployment/goagent -c goagent

# Previous instance (if crashed)
kubectl logs deployment/goagent --previous
```

Filter logs by level:

```bash
kubectl logs -l app=goagent | jq 'select(.level=="error")'
kubectl logs -l app=goagent | jq 'select(.level=="warn" or .level=="error")'
```

## Examples

### Kubernetes Monitoring Agent

Deploys a CronJob that monitors cluster health every 15 minutes:

```bash
kubectl apply -f examples/monitoring-agent.yaml
kubectl get cronjob monitoring-agent
kubectl logs -l app=monitoring-agent -f
```

Customize the monitoring task in ConfigMap:

```yaml
data:
  agent_task: |
    Monitor the Kubernetes cluster and report on:
    1. Pods that are not in Running state
    2. Services with no endpoints
    3. Deployments with unavailable replicas
```

### Incident Response Agent

Deploys a service that investigates incidents:

```bash
kubectl apply -f examples/incident-response.yaml
kubectl get deployment incident-agent
kubectl port-forward svc/incident-agent 8080:8080
```

Invoke via HTTP (when service mode implemented):

```bash
curl -X POST http://localhost:8080/investigate \
  -H "Content-Type: application/json" \
  -d '{"issue": "high memory usage in production namespace"}'
```

## Security Best Practices

### 1. Principle of Least Privilege

Only grant necessary RBAC permissions:

```bash
# Review current permissions
kubectl auth can-i --list --as=system:serviceaccount:default:goagent

# Test specific permission
kubectl auth can-i get pods --as=system:serviceaccount:default:goagent
```

### 2. Network Policies

Restrict network access:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: goagent-netpol
spec:
  podSelector:
    matchLabels:
      app: goagent
  policyTypes:
  - Egress
  egress:
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 443  # HTTPS for API calls
  - to:
    - podSelector:
        matchLabels:
          app: otel-collector
    ports:
    - protocol: TCP
      port: 4318  # OTLP
```

### 3. Pod Security Standards

Apply pod security standards:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: goagent
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

### 4. Secret Management

Use external secret management:

```bash
# Using External Secrets Operator
kubectl apply -f - <<EOF
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: goagent-secrets
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: vault-backend
    kind: SecretStore
  target:
    name: goagent-secrets
  data:
  - secretKey: openai_api_key
    remoteRef:
      key: secret/goagent/openai
      property: api_key
EOF
```

## Troubleshooting

### Pod Not Starting

Check pod status and events:

```bash
kubectl describe pod -l app=goagent
kubectl get events --sort-by='.lastTimestamp'
```

Common issues:
- **ImagePullBackOff**: Image not accessible or doesn't exist
- **CrashLoopBackOff**: Application crashing on startup
- **Pending**: Insufficient resources or scheduling issues

### RBAC Permission Denied

Test permissions:

```bash
# Check service account permissions
kubectl auth can-i get pods --as=system:serviceaccount:default:goagent -n default

# View role bindings
kubectl get rolebinding -o wide | grep goagent
```

Grant additional permissions if needed:

```bash
kubectl edit role goagent-reader
```

### High Memory Usage

Check resource usage:

```bash
kubectl top pod -l app=goagent
```

Increase limits:

```bash
kubectl set resources deployment goagent \
  --limits=cpu=2000m,memory=1Gi \
  --requests=cpu=200m,memory=256Mi
```

### CronJob Not Running

Check CronJob status:

```bash
kubectl get cronjob goagent-scheduled
kubectl describe cronjob goagent-scheduled
```

Manually trigger job:

```bash
kubectl create job --from=cronjob/goagent-scheduled goagent-manual
```

### Secrets Not Mounting

Verify secret exists:

```bash
kubectl get secret goagent-secrets
kubectl describe secret goagent-secrets
```

Check mount path in pod:

```bash
kubectl exec -it deployment/goagent -- ls -la /var/secrets/google
```

## Performance Tuning

### Optimization Tips

1. **Enable connection pooling** for provider API calls
2. **Adjust MAX_ITERATIONS** based on task complexity
3. **Set appropriate timeouts** to prevent hanging
4. **Use resource requests/limits** to ensure QoS
5. **Enable horizontal scaling** for high-load scenarios

### Monitoring Key Metrics

- Agent iteration count (should be < MAX_ITERATIONS)
- Token usage per request
- Request latency (p50, p95, p99)
- Error rate
- Memory usage trends

## Migration from Docker

Migrating from Docker deployment:

```bash
# 1. Build and push image to registry
docker build -t your-registry.io/goagent:v1.0.0 .
docker push your-registry.io/goagent:v1.0.0

# 2. Update image in manifests
sed -i 's|goagent:latest|your-registry.io/goagent:v1.0.0|g' base/*.yaml

# 3. Create secrets from Docker .env
kubectl create secret generic goagent-secrets \
  --from-env-file=.env

# 4. Deploy to cluster
kubectl apply -f base/
```

## Next Steps

- Set up CI/CD pipeline for automated deployments
- Configure alerting for agent failures
- Implement custom tools for your use case
- Explore event-driven patterns with KEDA
- Set up multi-cluster deployment with GitOps

## Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator)
- [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/)
- [KEDA (Event-Driven Autoscaling)](https://keda.sh/)
- [Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets)
