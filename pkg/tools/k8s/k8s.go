// Package k8s provides a built-in Kubernetes client tool for agents.
// It supports querying cluster resources with RBAC-safe operations.
package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/oskarhane/goagent/pkg/tools"
	"github.com/oskarhane/goagent/pkg/types"
)

const (
	operationGet  = "get"
	operationList = "list"
	namespaceAll  = "all"
)

// Config configures the Kubernetes tool behavior.
type Config struct {
	// KubeconfigPath specifies the path to the kubeconfig file.
	// If empty, uses default behavior (KUBECONFIG env, ~/.kube/config, or in-cluster config).
	KubeconfigPath string

	// DefaultTimeout is the default timeout for Kubernetes operations.
	// If not set, defaults to 30 seconds.
	DefaultTimeout time.Duration

	// DefaultNamespace is the default namespace for resource queries.
	// If empty, uses "default".
	DefaultNamespace string
}

// Params defines the parameters for Kubernetes queries.
type Params struct {
	Operation string `json:"operation"` // get, list, describe
	Resource  string `json:"resource"`  // pod, service, deployment, etc.
	Name      string `json:"name"`      // resource name (optional for list)
	Namespace string `json:"namespace"` // namespace (optional, defaults to "default")
	Labels    string `json:"labels"`    // label selector (optional, e.g., "app=myapp")
	Timeout   int    `json:"timeout"`   // in seconds
}

// Response represents the result of a Kubernetes query.
type Response struct {
	Operation string `json:"operation"`
	Resource  string `json:"resource"`
	Namespace string `json:"namespace"`
	Data      any    `json:"data"`
	Error     string `json:"error,omitempty"`
}

// NewTool creates a new Kubernetes tool definition.
// The tool supports querying common resources like pods, services, and deployments.
func NewTool() types.Tool {
	return tools.NewBuilder(
		"k8s_query",
		"Query Kubernetes cluster resources. "+
			"Supports common operations (get, list) for pods, services, deployments, and other resources. "+
			"RBAC permissions apply - only resources the service account can access are returned.",
	).
		StringParamWithEnum(
			"operation",
			"Operation to perform",
			true,
			[]string{operationGet, operationList},
		).
		StringParamWithEnum(
			"resource",
			"Type of Kubernetes resource",
			true,
			[]string{
				"pod", "pods", "service", "services",
				"deployment", "deployments", "configmap", "configmaps",
				"secret", "secrets", "namespace", "namespaces", "node", "nodes",
			},
		).
		StringParam(
			"name",
			"Name of the specific resource (required for 'get' operation)",
			false,
		).
		StringParam(
			"namespace",
			"Namespace to query (defaults to 'default', use 'all' for cluster-wide list)",
			false,
		).
		StringParam(
			"labels",
			"Label selector for filtering (e.g., 'app=myapp,env=prod')",
			false,
		).
		IntegerParam("timeout", "Optional timeout in seconds (default: 30, max: 300)", false).
		Build()
}

// NewHandler creates a new Kubernetes tool handler with the given configuration.
// If config is nil, default values are used.
func NewHandler(config *Config) tools.Handler {
	if config == nil {
		config = &Config{}
	}
	if config.DefaultTimeout == 0 {
		config.DefaultTimeout = 30 * time.Second
	}
	if config.DefaultNamespace == "" {
		config.DefaultNamespace = "default"
	}

	return func(ctx context.Context, call types.ToolCall) types.ToolResult {
		start := time.Now()

		// Parse parameters
		var params Params
		if err := types.ParseToolArguments(call, &params); err != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("invalid parameters: %v", err),
				ExecutionTime: time.Since(start),
			}
		}

		// Validate operation
		params.Operation = strings.ToLower(params.Operation)
		if params.Operation != operationGet && params.Operation != operationList {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("unsupported operation: %s", params.Operation),
				ExecutionTime: time.Since(start),
			}
		}

		// Validate resource
		params.Resource = strings.ToLower(params.Resource)

		// For "get" operation, name is required
		if params.Operation == operationGet && strings.TrimSpace(params.Name) == "" {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         "name is required for 'get' operation",
				ExecutionTime: time.Since(start),
			}
		}

		// Set default namespace
		if params.Namespace == "" {
			params.Namespace = config.DefaultNamespace
		}

		// Determine timeout
		timeout := config.DefaultTimeout
		if params.Timeout > 0 {
			if params.Timeout > 300 {
				params.Timeout = 300 // max 5 minutes
			}
			timeout = time.Duration(params.Timeout) * time.Second
		}

		// Create Kubernetes client context with timeout
		k8sCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Build Kubernetes client config
		var k8sConfig *rest.Config
		var err error

		if config.KubeconfigPath != "" {
			// Use specified kubeconfig path
			k8sConfig, err = clientcmd.BuildConfigFromFlags("", config.KubeconfigPath)
		} else {
			// Try default kubeconfig loading rules (KUBECONFIG env, ~/.kube/config)
			loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
			configOverrides := &clientcmd.ConfigOverrides{}
			k8sConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				loadingRules,
				configOverrides,
			).ClientConfig()

			// If kubeconfig loading fails, try in-cluster config
			if err != nil {
				k8sConfig, err = rest.InClusterConfig()
			}
		}

		if err != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("failed to build Kubernetes config: %v", err),
				ExecutionTime: time.Since(start),
			}
		}

		// Create Kubernetes clientset
		clientset, err := kubernetes.NewForConfig(k8sConfig)
		if err != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("failed to create Kubernetes client: %v", err),
				ExecutionTime: time.Since(start),
			}
		}

		// Execute the query
		var data any
		switch params.Resource {
		case "pod", "pods":
			data, err = queryPods(k8sCtx, clientset, &params)
		case "service", "services":
			data, err = queryServices(k8sCtx, clientset, &params)
		case "deployment", "deployments":
			data, err = queryDeployments(k8sCtx, clientset, &params)
		case "configmap", "configmaps":
			data, err = queryConfigMaps(k8sCtx, clientset, &params)
		case "secret", "secrets":
			data, err = querySecrets(k8sCtx, clientset, &params)
		case "namespace", "namespaces":
			data, err = queryNamespaces(k8sCtx, clientset, &params)
		case "node", "nodes":
			data, err = queryNodes(k8sCtx, clientset, &params)
		default:
			err = fmt.Errorf("unsupported resource type: %s", params.Resource)
		}

		// Check for context cancellation
		if k8sCtx.Err() != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("operation timeout after %v", timeout),
				ExecutionTime: time.Since(start),
			}
		}

		// Create response
		resp := Response{
			Operation: params.Operation,
			Resource:  params.Resource,
			Namespace: params.Namespace,
			Data:      data,
		}

		if err != nil {
			resp.Error = err.Error()
		}

		// Marshal response to JSON
		resultJSON, err := json.Marshal(resp)
		if err != nil {
			return types.ToolResult{
				ToolCallID:    call.ID,
				ToolName:      call.Function.Name,
				Error:         fmt.Sprintf("failed to marshal response: %v", err),
				ExecutionTime: time.Since(start),
			}
		}

		return types.ToolResult{
			ToolCallID:    call.ID,
			ToolName:      call.Function.Name,
			Content:       string(resultJSON),
			ExecutionTime: time.Since(start),
		}
	}
}

// queryPods queries pod resources
func queryPods(ctx context.Context, clientset *kubernetes.Clientset, params *Params) (any, error) {
	opts := metav1.ListOptions{
		LabelSelector: params.Labels,
	}

	if params.Operation == operationGet {
		return clientset.CoreV1().Pods(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
	}

	// Handle cluster-wide list
	namespace := params.Namespace
	if namespace == namespaceAll {
		namespace = metav1.NamespaceAll
	}

	return clientset.CoreV1().Pods(namespace).List(ctx, opts)
}

// queryServices queries service resources
func queryServices(ctx context.Context, clientset *kubernetes.Clientset, params *Params) (any, error) {
	opts := metav1.ListOptions{
		LabelSelector: params.Labels,
	}

	if params.Operation == operationGet {
		return clientset.CoreV1().Services(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
	}

	namespace := params.Namespace
	if namespace == namespaceAll {
		namespace = metav1.NamespaceAll
	}

	return clientset.CoreV1().Services(namespace).List(ctx, opts)
}

// queryDeployments queries deployment resources
func queryDeployments(ctx context.Context, clientset *kubernetes.Clientset, params *Params) (any, error) {
	opts := metav1.ListOptions{
		LabelSelector: params.Labels,
	}

	if params.Operation == operationGet {
		return clientset.AppsV1().Deployments(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
	}

	namespace := params.Namespace
	if namespace == namespaceAll {
		namespace = metav1.NamespaceAll
	}

	return clientset.AppsV1().Deployments(namespace).List(ctx, opts)
}

// queryConfigMaps queries configmap resources
func queryConfigMaps(ctx context.Context, clientset *kubernetes.Clientset, params *Params) (any, error) {
	opts := metav1.ListOptions{
		LabelSelector: params.Labels,
	}

	if params.Operation == operationGet {
		return clientset.CoreV1().ConfigMaps(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
	}

	namespace := params.Namespace
	if namespace == namespaceAll {
		namespace = metav1.NamespaceAll
	}

	return clientset.CoreV1().ConfigMaps(namespace).List(ctx, opts)
}

// querySecrets queries secret resources
func querySecrets(ctx context.Context, clientset *kubernetes.Clientset, params *Params) (any, error) {
	opts := metav1.ListOptions{
		LabelSelector: params.Labels,
	}

	if params.Operation == operationGet {
		return clientset.CoreV1().Secrets(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
	}

	namespace := params.Namespace
	if namespace == namespaceAll {
		namespace = metav1.NamespaceAll
	}

	return clientset.CoreV1().Secrets(namespace).List(ctx, opts)
}

// queryNamespaces queries namespace resources
func queryNamespaces(ctx context.Context, clientset *kubernetes.Clientset, params *Params) (any, error) {
	opts := metav1.ListOptions{
		LabelSelector: params.Labels,
	}

	if params.Operation == operationGet {
		return clientset.CoreV1().Namespaces().Get(ctx, params.Name, metav1.GetOptions{})
	}

	return clientset.CoreV1().Namespaces().List(ctx, opts)
}

// queryNodes queries node resources
func queryNodes(ctx context.Context, clientset *kubernetes.Clientset, params *Params) (any, error) {
	opts := metav1.ListOptions{
		LabelSelector: params.Labels,
	}

	if params.Operation == operationGet {
		return clientset.CoreV1().Nodes().Get(ctx, params.Name, metav1.GetOptions{})
	}

	return clientset.CoreV1().Nodes().List(ctx, opts)
}
