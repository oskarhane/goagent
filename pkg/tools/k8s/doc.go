// Package k8s provides a built-in Kubernetes client tool for querying cluster resources.
//
// The Kubernetes tool allows agents to interact with Kubernetes clusters by querying
// resources like pods, services, deployments, and other common workload types.
// All operations are RBAC-safe and respect the service account permissions.
//
// # Features
//
//   - Query operations: get (single resource) and list (multiple resources)
//   - Support for common resources: pods, services, deployments, configmaps, secrets, namespaces, nodes
//   - Namespace scoping with cluster-wide list support
//   - Label selector filtering for targeted queries
//   - Configurable timeouts with context cancellation
//   - Automatic kubeconfig detection (KUBECONFIG env, ~/.kube/config, or in-cluster)
//   - RBAC-safe operations (only returns what the service account can access)
//
// # Authentication
//
// The tool uses the standard Kubernetes authentication chain:
//
//  1. If KubeconfigPath is set in Config, uses that specific kubeconfig file
//  2. Otherwise, checks KUBECONFIG environment variable
//  3. Falls back to ~/.kube/config
//  4. Finally tries in-cluster configuration (for pods running in Kubernetes)
//
// # Usage
//
// Basic usage with default configuration:
//
//	registry := tools.NewRegistry()
//	k8sTool := k8s.NewTool()
//	k8sHandler := k8s.NewHandler(nil) // Uses defaults
//	registry.MustRegister(k8sTool, k8sHandler)
//
// Custom configuration with specific kubeconfig:
//
//	config := &k8s.Config{
//	    KubeconfigPath:   "/path/to/kubeconfig",
//	    DefaultTimeout:   60 * time.Second,
//	    DefaultNamespace: "production",
//	}
//	k8sHandler := k8s.NewHandler(config)
//	registry.MustRegister(k8s.NewTool(), k8sHandler)
//
// # Example Agent Usage
//
// Get a specific pod:
//
//	{
//	    "operation": "get",
//	    "resource": "pod",
//	    "name": "nginx-abc123",
//	    "namespace": "default"
//	}
//
// List all pods in a namespace with label selector:
//
//	{
//	    "operation": "list",
//	    "resource": "pods",
//	    "namespace": "production",
//	    "labels": "app=nginx,env=prod"
//	}
//
// List all services across all namespaces:
//
//	{
//	    "operation": "list",
//	    "resource": "services",
//	    "namespace": "all"
//	}
//
// Get deployment status:
//
//	{
//	    "operation": "get",
//	    "resource": "deployment",
//	    "name": "web-app",
//	    "namespace": "default"
//	}
//
// # Response Format
//
// The tool returns a JSON response with the following structure:
//
//	{
//	    "operation": "get",
//	    "resource": "pod",
//	    "namespace": "default",
//	    "data": {
//	        "metadata": {"name": "nginx-abc123", "namespace": "default"},
//	        "spec": {...},
//	        "status": {"phase": "Running", ...}
//	    },
//	    "error": ""  // Only present if operation failed
//	}
//
// For list operations, the data field contains the full list response from the Kubernetes API.
//
// # Supported Resources
//
//   - Pods (pod/pods)
//   - Services (service/services)
//   - Deployments (deployment/deployments)
//   - ConfigMaps (configmap/configmaps)
//   - Secrets (secret/secrets)
//   - Namespaces (namespace/namespaces)
//   - Nodes (node/nodes)
//
// # Security Considerations
//
//   - All operations respect RBAC permissions of the kubeconfig/service account
//   - Use narrow service account permissions in production environments
//   - Set appropriate timeouts to prevent hanging operations
//   - Be cautious when granting access to secrets
//   - Cluster-wide operations (namespace: "all") require broader permissions
//
// # Common Use Cases
//
//   - Monitoring pod status and health checks
//   - Querying service endpoints for troubleshooting
//   - Checking deployment rollout status
//   - Retrieving configuration for debugging
//   - Listing resources by label for inventory management
//   - Node capacity and health monitoring
//
// # RBAC Configuration
//
// For in-cluster usage, create a ServiceAccount with appropriate permissions:
//
//	apiVersion: v1
//	kind: ServiceAccount
//	metadata:
//	  name: goagent-sa
//	  namespace: default
//	---
//	apiVersion: rbac.authorization.k8s.io/v1
//	kind: Role
//	metadata:
//	  name: goagent-reader
//	  namespace: default
//	rules:
//	- apiGroups: [""]
//	  resources: ["pods", "services", "configmaps"]
//	  verbs: ["get", "list"]
//	- apiGroups: ["apps"]
//	  resources: ["deployments"]
//	  verbs: ["get", "list"]
//	---
//	apiVersion: rbac.authorization.k8s.io/v1
//	kind: RoleBinding
//	metadata:
//	  name: goagent-reader-binding
//	  namespace: default
//	roleRef:
//	  apiGroup: rbac.authorization.k8s.io
//	  kind: Role
//	  name: goagent-reader
//	subjects:
//	- kind: ServiceAccount
//	  name: goagent-sa
//	  namespace: default
package k8s
