# Deployment

DEPLOYMENT STRATEGY: containerized Go service/job, deploy via Docker Compose or Kubernetes manifests; release via GoReleaser
CONTAINERIZATION: multi-stage Dockerfiles for base, service, cronjob images; Compose examples for OpenAI/Vertex
CI/CD: GitHub Actions CI + Docker build/push to GHCR + GoReleaser on tags
HOSTING: generic containers on Kubernetes or Docker runtime; no platform-specific configs
ENVIRONMENT MANAGEMENT: `.env` examples, docker-compose env vars, K8s ConfigMap + Secret

**Details**
- **Container orchestration:** Kubernetes manifests in `deployments/k8s/base/*` plus examples; Docker Compose in `deployments/docker/docker-compose.yml`.
- **Cloud platforms:** no AWS/GCP/Azure/Vercel/Netlify/Railway configs detected; images published to GHCR via Actions.
- **IaC:** no Terraform/CloudFormation/Pulumi found.
- **Serverless/static:** no Lambda/Cloud Functions/Vercel/Netlify deployment files; not a static site.
- **Database/migrations:** no DB configs or migration tooling detected.
- **Env config:** `.env.example` files in `deployments/docker/.env.example` and `examples/*/.env.example`; K8s `ConfigMap` + `Secret` templates.
- **Monitoring/logging:** OpenTelemetry collector example `deployments/k8s/monitoring/otel-collector.yaml`; Prometheus `ServiceMonitor`/`PodMonitor`; container healthchecks in Dockerfiles.

If you want, I can map a concrete deployment plan for a target platform (EKS/GKE/AKS/EC2) or add IaC/templates.

---

*This file is part of the AGENTS.md documentation system.*
