package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/oskarhane/goagent/pkg/agent"
	"github.com/oskarhane/goagent/pkg/logger"
	"github.com/oskarhane/goagent/pkg/providers/openai"
	"github.com/oskarhane/goagent/pkg/tools"
	"github.com/oskarhane/goagent/pkg/types"

	_ "github.com/joho/godotenv/autoload"
)

// Agent-as-Tool Pattern for SRE Incident Response
// This example demonstrates hierarchical agent delegation:
// - Coordinator agent: orchestrates investigation, delegates to service investigators
// - Investigator agents: wrapped as tools, perform focused service-level diagnostics

func main() {
	// Get API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("OPENAI_API_KEY environment variable is required")
	}

	// Create OpenAI provider with gpt-5.1 for complex reasoning
	provider, err := openai.NewProvider(&openai.Config{
		APIKey: apiKey,
		Model:  "gpt-5.1",
	})
	if err != nil {
		log.Fatalf("Failed to create OpenAI provider: %v", err)
	}

	// Create logger with debug mode support
	logLevel := logger.LevelInfo
	if os.Getenv("DEBUG") == "true" {
		logLevel = logger.LevelDebug
	}
	l := logger.New(logger.Config{
		Level:  logLevel,
		Output: os.Stdout,
	})

	// Create mock tools registry for investigator agents
	// These tools provide simulated diagnostic data (logs, metrics, status)
	mockRegistry := tools.NewRegistry()
	mockRegistry.MustRegister(NewMockLogsTool(), NewMockLogsHandler())
	mockRegistry.MustRegister(NewMockMetricsTool(), NewMockMetricsHandler())
	mockRegistry.MustRegister(NewMockServiceStatusTool(), NewMockServiceStatusHandler())

	// Create coordinator tools registry
	// The coordinator only has access to investigate_service, which wraps investigator agents
	coordinatorRegistry := tools.NewRegistry()
	investigateTool := NewInvestigateServiceTool()
	investigateHandler := NewInvestigateServiceHandler(provider, mockRegistry)
	coordinatorRegistry.MustRegister(investigateTool, investigateHandler)

	// Create coordinator agent with SRE triage system prompt
	coordinatorPrompt := `You are an SRE coordinator responsible for incident triage and investigation orchestration.

Your role is to:
1. Analyze incident reports to understand symptoms and scope
2. Identify which services need investigation based on the incident type
3. Delegate investigations to service-specific agents using the investigate_service tool
4. Correlate findings from multiple service investigations
5. Determine the root cause and provide actionable recommendations

You have access to:
- investigate_service: Delegates investigation to a specialized agent that can examine logs, metrics, and health status for a specific service

Investigation strategy:
- For cascading failures: investigate suspected root cause services first (e.g., database, auth-service)
- For isolated issues: focus on directly affected services
- Always investigate at least 2-3 services to understand the full scope
- Look for patterns across service investigations (timeouts, errors, resource exhaustion)

Be systematic and thorough. Provide a clear incident summary with:
- Root cause identification
- Services affected and their status
- Evidence supporting your conclusion
- Recommended remediation steps`

	coordinatorConfig := &agent.Config{
		Provider:      provider,
		SystemPrompt:  coordinatorPrompt,
		Registry:      coordinatorRegistry,
		MaxIterations: 10,
		Logger:        l,
	}

	coordinator, err := agent.NewAgent(coordinatorConfig)
	if err != nil {
		log.Fatalf("Failed to create coordinator agent: %v", err)
	}

	// Get incident description from command line args or use default scenario
	var incident string
	if len(os.Args) > 1 {
		incident = os.Args[1]
	} else {
		// Use default scenario
		scenarios := NewScenarios()
		defaultScenario := scenarios.Default()
		incident = defaultScenario.GetIncidentDescription()
	}

	// Run coordinator agent with 60s timeout to prevent runaway execution
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("=== Agent-as-Tool: Hierarchical SRE Investigation ===")
	fmt.Println("\n--- Incident Report ---")
	fmt.Println(incident)
	fmt.Println()

	// Execute investigation
	result := coordinator.Run(ctx, incident, nil)

	// Check for errors
	if result.Error != nil {
		log.Fatalf("Investigation failed: %v", result.Error)
	}

	// Print delegation hierarchy - show which services were investigated
	fmt.Println("\n--- Delegation Hierarchy ---")
	fmt.Println("Coordinator → Investigator Agents:")
	investigatedServices := extractInvestigatedServices(result.Messages)
	if len(investigatedServices) == 0 {
		fmt.Println("  (No delegations occurred)")
	} else {
		for i, service := range investigatedServices {
			fmt.Printf("  %d. Investigating: %s\n", i+1, service)
		}
	}

	// Print investigation findings from delegated agents
	fmt.Println("\n--- Investigation Findings ---")
	findings := extractInvestigationFindings(result.Messages)
	if len(findings) == 0 {
		fmt.Println("(No investigation findings)")
	} else {
		for i, finding := range findings {
			fmt.Printf("\nService Investigation #%d:\n%s\n", i+1, finding)
			fmt.Println("---")
		}
	}

	// Print final coordinator analysis
	fmt.Println("\n=== Coordinator Final Analysis ===")
	fmt.Println(result.Response.Content)

	// Print execution statistics
	fmt.Printf("\n=== Execution Statistics ===\n")
	fmt.Printf("Coordinator iterations: %d\n", result.Iterations)
	fmt.Printf("Total tokens used: %d\n", result.TotalTokens)
	fmt.Printf("Total execution time: %.2fs\n", result.ExecutionTime.Seconds())
}

// extractInvestigatedServices parses messages to find which services were investigated
func extractInvestigatedServices(messages []types.Message) []string {
	var services []string
	for _, msg := range messages {
		if msg.Role == types.RoleAssistant && len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				if tc.Function.Name == "investigate_service" {
					// Parse the service name from arguments
					var args struct {
						ServiceName string `json:"service_name"`
					}
					if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err == nil {
						services = append(services, args.ServiceName)
					}
				}
			}
		}
	}
	return services
}

// extractInvestigationFindings extracts tool results from investigate_service calls
func extractInvestigationFindings(messages []types.Message) []string {
	var findings []string
	for _, msg := range messages {
		if msg.Role == types.RoleTool && msg.Name == "investigate_service" {
			findings = append(findings, msg.Content)
		}
	}
	return findings
}
