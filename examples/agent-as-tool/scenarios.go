package main

import "fmt"

// Scenario represents a pre-built incident scenario for demonstration
type Scenario struct {
	Name             string
	IncidentType     string
	InitialReport    string
	AffectedServices []string
	RootCause        string
	Description      string
}

// Scenarios holds all available incident scenarios
type Scenarios struct {
	scenarios []Scenario
}

// NewScenarios creates and returns all available incident scenarios
func NewScenarios() *Scenarios {
	return &Scenarios{
		scenarios: []Scenario{
			{
				Name:             "Database Cascading Failure",
				IncidentType:     "cascading_failure",
				InitialReport:    "Multiple services reporting degraded performance. Users experiencing login failures and API timeouts. Started approximately 15 minutes ago with increased error rates on authentication endpoints.",
				AffectedServices: []string{"database", "auth-service", "api-service"},
				RootCause:        "Database connection pool exhaustion causing auth-service timeouts, leading to API service degradation",
				Description:      "A classic cascading failure where database performance issues (slow queries, connection pool saturation) cause the auth-service to timeout, which then causes the api-service to fail authentication checks, ultimately impacting end users.",
			},
			{
				Name:             "Memory Leak in Cache Service",
				IncidentType:     "memory_leak",
				InitialReport:    "Cache service showing steadily increasing memory usage over the past 2 hours. Hit rates remain normal but response times are degrading. No deployments in the last 24 hours.",
				AffectedServices: []string{"cache"},
				RootCause:        "Memory leak in cache service causing gradual performance degradation",
				Description:      "A memory leak scenario where the cache service slowly consumes more memory, eventually leading to performance degradation as the system begins swapping or the OOM killer activates.",
			},
			{
				Name:             "Auth Service Database Dependency Failure",
				IncidentType:     "dependency_failure",
				InitialReport:    "Authentication failures spiking. Users reporting 'Service Unavailable' errors on login. Database metrics showing normal operation. Issue started 10 minutes ago.",
				AffectedServices: []string{"auth-service", "api-service"},
				RootCause:        "Auth-service unable to connect to database despite database operating normally (network partition or misconfiguration)",
				Description:      "An isolated dependency failure where auth-service loses connectivity to the database even though the database itself is healthy, demonstrating the importance of checking dependency health separately from service health.",
			},
		},
	}
}

// Get returns a scenario by index
func (s *Scenarios) Get(index int) (Scenario, error) {
	if index < 0 || index >= len(s.scenarios) {
		return Scenario{}, fmt.Errorf("scenario index %d out of range (0-%d)", index, len(s.scenarios)-1)
	}
	return s.scenarios[index], nil
}

// GetByName returns a scenario by name
func (s *Scenarios) GetByName(name string) (Scenario, error) {
	for _, scenario := range s.scenarios {
		if scenario.Name == name {
			return scenario, nil
		}
	}
	return Scenario{}, fmt.Errorf("scenario '%s' not found", name)
}

// List returns all available scenarios
func (s *Scenarios) List() []Scenario {
	return s.scenarios
}

// Default returns the default scenario (first one - cascading failure)
func (s *Scenarios) Default() Scenario {
	if len(s.scenarios) == 0 {
		return Scenario{}
	}
	return s.scenarios[0]
}

// GetIncidentDescription formats the initial incident report for the coordinator agent
func (s *Scenario) GetIncidentDescription() string {
	return fmt.Sprintf("INCIDENT REPORT:\n\nType: %s\nReport: %s\n\nYour task is to investigate this incident, determine the root cause, and provide recommendations for resolution.",
		s.IncidentType,
		s.InitialReport,
	)
}

// GetExpectedServices returns the list of services that should be investigated
func (s *Scenario) GetExpectedServices() []string {
	return s.AffectedServices
}
