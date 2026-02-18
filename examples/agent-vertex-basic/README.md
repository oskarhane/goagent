# Agent Vertex Basic

This example mirrors `examples/agent-basic` but uses the Vertex AI provider.

## Requirements

- `VERTEX_PROJECT_ID` (required)
- `VERTEX_LOCATION` (optional, defaults to provider default)
- `VERTEX_MODEL` (optional, defaults to provider default)

Authentication uses Google Cloud Application Default Credentials (ADC). Configure one of:

- `GOOGLE_APPLICATION_CREDENTIALS` pointing to a service account key file
- Workload Identity (GKE)
- Default credentials when running on Google Cloud
