package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/adafia/solid-fortnight/internal/config"
)

func TestGatewayIntegration(t *testing.T) {
	// 1. Set up mock backends
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from backend: %s", r.URL.Path)
	}))
	defer backend.Close()

	// 2. Set environment variables to point gateway to mock backend
	os.Setenv("MANAGEMENT_HOST", "localhost")
	os.Setenv("EVALUATOR_HOST", "localhost")
	os.Setenv("STREAMER_HOST", "localhost")
	os.Setenv("ANALYTICS_HOST", "localhost")

	// 3. Create a config that uses the backend ports
	cfg := &config.Config{
		Services: map[string]config.ServiceConfig{
			"gateway":    {Port: 8080},
			"management": {Port: getPort(backend.URL)},
			"evaluator":  {Port: getPort(backend.URL)},
			"streamer":   {Port: getPort(backend.URL)},
			"analytics":  {Port: getPort(backend.URL)},
		},
	}

	handler := NewGatewayHandler(cfg)
	gatewayServer := httptest.NewServer(handler)
	defer gatewayServer.Close()

	// 4. Test various proxy paths
	tests := []struct {
		name           string
		path           string
		expectedBody   string
		expectedStatus int
	}{
		{
			name:           "Management Proxy (path mapping)",
			path:           "/api/v1/management/projects",
			expectedBody:   "Hello from backend: /projects",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Evaluator Proxy (no mapping)",
			path:           "/api/v1/evaluate",
			expectedBody:   "Hello from backend: /api/v1/evaluate",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Streamer Proxy (mapping)",
			path:           "/api/v1/stream",
			expectedBody:   "Hello from backend: /stream",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Analytics Proxy (mapping)",
			path:           "/api/v1/analytics/events",
			expectedBody:   "Hello from backend: /api/v1/events",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Not Found",
			path:           "/api/v1/non-existent",
			expectedBody:   "404 page not found\n",
			expectedStatus: http.StatusNotFound,
		},
	}

	client := gatewayServer.Client()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Get(gatewayServer.URL + tt.path)
			if err != nil {
				t.Fatalf("Failed to request gateway: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			body, _ := io.ReadAll(resp.Body)
			if string(body) != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, string(body))
			}
		})
	}
}

func getPort(u string) int {
	var port int
	fmt.Sscanf(u[strings.LastIndex(u, ":")+1:], "%d", &port)
	return port
}
