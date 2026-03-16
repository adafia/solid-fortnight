package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/adafia/solid-fortnight/apps/gateway/middleware"
	"github.com/adafia/solid-fortnight/internal/config"
)

func main() {
	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "deployments/config.yaml"
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration from %s: %v", configPath, err)
	}

	handler := NewGatewayHandler(cfg)

	port := cfg.Services["gateway"].Port
	if port == 0 {
		port = 8080
	}
	addr := fmt.Sprintf(":%d", port)
	log.Printf("API Gateway listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Failed to start gateway: %v", err)
	}
}

func NewGatewayHandler(cfg *config.Config) http.Handler {
	// Service discovery (individual overrides or defaults)
	targets := map[string]string{
		"management": getServiceURL("management", "localhost", cfg.Services["management"].Port),
		"evaluator":  getServiceURL("evaluator", "localhost", cfg.Services["evaluator"].Port),
		"streamer":   getServiceURL("streamer", "localhost", cfg.Services["streamer"].Port),
		"analytics":  getServiceURL("analytics", "localhost", cfg.Services["analytics"].Port),
	}

	mux := http.NewServeMux()

	// Reverse proxies with target parsing done once
	// /api/v1/management/projects/ -> /projects/
	mux.Handle("/api/v1/management/", middleware.Chain(createProxyHandler(targets["management"], "/api/v1/management", ""), middleware.Logger, middleware.Auth, middleware.RateLimit))
	// /api/v1/evaluate -> /api/v1/evaluate
	mux.Handle("/api/v1/evaluate", middleware.Chain(createProxyHandler(targets["evaluator"], "", ""), middleware.Logger, middleware.Auth, middleware.RateLimit))
	// /api/v1/stream -> /stream
	mux.Handle("/api/v1/stream", middleware.Chain(createProxyHandler(targets["streamer"], "/api/v1", ""), middleware.Logger, middleware.Auth, middleware.RateLimit))
	// /api/v1/analytics/events/ -> /api/v1/events/
	mux.Handle("/api/v1/analytics/", middleware.Chain(createProxyHandler(targets["analytics"], "/analytics", ""), middleware.Logger, middleware.Auth, middleware.RateLimit))

	return mux
}

func getServiceURL(serviceName, defaultHost string, port int) string {
	host := os.Getenv(strings.ToUpper(serviceName) + "_HOST")
	if host == "" {
		host = defaultHost
	}
	return fmt.Sprintf("http://%s:%d", host, port)
}

func createProxyHandler(target string, prefixToReplace, replacement string) http.HandlerFunc {
	u, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Failed to parse target URL %s: %v", target, err)
	}
	proxy := httputil.NewSingleHostReverseProxy(u)

	// Custom director to handle path mapping and Host header
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		if prefixToReplace != "" {
			req.URL.Path = strings.Replace(req.URL.Path, prefixToReplace, replacement, 1)
		}
		// Ensure the Host header matches the target for some backends that care
		req.Host = u.Host
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Proxying request: %s %s -> %s", r.Method, r.URL.Path, target)
		proxy.ServeHTTP(w, r)
	}
}
