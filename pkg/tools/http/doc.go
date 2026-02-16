// Package http provides a built-in HTTP client tool for making REST API calls.
//
// The HTTP tool allows agents to interact with external web services and APIs
// by making HTTP requests with configurable methods, headers, and request bodies.
//
// # Features
//
//   - Support for GET, POST, PUT, DELETE methods
//   - Custom authentication headers (Bearer tokens, API keys, etc.)
//   - Request and response body handling
//   - Configurable timeouts with context cancellation
//   - Automatic Content-Type detection for JSON bodies
//   - Response size limiting to prevent memory issues
//   - Structured error handling for non-2xx responses
//
// # Usage
//
// Basic usage with default configuration:
//
//	registry := tools.NewRegistry()
//	httpTool := http.NewTool()
//	httpHandler := http.NewHandler(nil) // Uses defaults
//	registry.MustRegister(httpTool, httpHandler)
//
// Custom configuration:
//
//	config := &http.Config{
//	    DefaultTimeout:  60 * time.Second,
//	    MaxResponseSize: 5 * 1024 * 1024, // 5MB
//	}
//	httpHandler := http.NewHandler(config)
//	registry.MustRegister(http.NewTool(), httpHandler)
//
// # Example Agent Usage
//
// An agent can use this tool by calling it with parameters:
//
//	{
//	    "url": "https://api.github.com/users/octocat",
//	    "method": "GET",
//	    "headers": {
//	        "Accept": "application/json",
//	        "User-Agent": "GoAgent/1.0"
//	    }
//	}
//
// POST request with authentication:
//
//	{
//	    "url": "https://api.example.com/data",
//	    "method": "POST",
//	    "headers": {
//	        "Authorization": "Bearer your-token-here",
//	        "Content-Type": "application/json"
//	    },
//	    "body": "{\"key\": \"value\"}",
//	    "timeout": 60
//	}
//
// # Response Format
//
// The tool returns a JSON response with the following structure:
//
//	{
//	    "status_code": 200,
//	    "status": "200 OK",
//	    "headers": {
//	        "Content-Type": "application/json",
//	        "Date": "Mon, 16 Feb 2026 12:00:00 GMT"
//	    },
//	    "body": "{\"login\": \"octocat\", \"id\": 1}",
//	    "error": ""  // Only present for 4xx/5xx responses
//	}
//
// # Security Considerations
//
//   - Always use HTTPS for sensitive data
//   - Store authentication tokens in environment variables, not in code
//   - Set appropriate timeouts to prevent hanging requests
//   - Be aware of rate limits imposed by target APIs
//   - Response size is limited to prevent memory exhaustion
//
// # Common Use Cases
//
//   - Querying REST APIs for real-time data
//   - Posting alerts to webhook endpoints
//   - Retrieving configuration from external services
//   - Interacting with cloud provider APIs
//   - Fetching status from monitoring endpoints
package http
