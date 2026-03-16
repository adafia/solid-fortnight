package sdk

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/adafia/solid-fortnight/internal/engine"
)

type Config struct {
	EvaluatorURL  string
	StreamerURL   string
	EnvironmentID string
	PollInterval  time.Duration
}

type Client struct {
	config    Config
	evaluator *engine.Evaluator
	flags     map[string]engine.FlagConfig
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewClient(config Config) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Client{
		config:    config,
		evaluator: engine.NewEvaluator(),
		flags:     make(map[string]engine.FlagConfig),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Initial fetch
	if err := c.fetchFlags(); err != nil {
		log.Printf("Warning: initial flag fetch failed: %v", err)
	}

	// Start background synchronization
	go c.syncLoop()

	return c, nil
}

func (c *Client) Close() {
	c.cancel()
}

func (c *Client) fetchFlags() error {
	url := fmt.Sprintf("%s/api/v1/evaluate?environment_id=%s", c.config.EvaluatorURL, c.config.EnvironmentID)
	// We use the GET /api/v1/evaluate which we mapped to GetFlags in the handler
	
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch flags: status %d", resp.StatusCode)
	}

	var configs []engine.FlagConfig
	if err := json.NewDecoder(resp.Body).Decode(&configs); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, f := range configs {
		c.flags[f.Key] = f
	}

	return nil
}

func (c *Client) syncLoop() {
	// 1. SSE Stream
	go c.streamUpdates()

	// 2. Periodic polling as fallback
	if c.config.PollInterval == 0 {
		c.config.PollInterval = 5 * time.Minute
	}
	ticker := time.NewTicker(c.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.fetchFlags(); err != nil {
				log.Printf("Polling error: %v", err)
			}
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Client) streamUpdates() {
	streamURL := fmt.Sprintf("%s/stream?environment_id=%s", c.config.StreamerURL, c.config.EnvironmentID)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			if err := c.connectStream(streamURL); err != nil {
				log.Printf("SSE connection error: %v. Retrying in 5 seconds...", err)
				time.Sleep(5 * time.Second)
			}
		}
	}
}

func (c *Client) connectStream(url string) error {
	req, err := http.NewRequestWithContext(c.ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("SSE status %d", resp.StatusCode)
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil // Connection closed by server
			}
			return err
		}

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data: update") {
			log.Println("Received update event from streamer, fetching flags...")
			if err := c.fetchFlags(); err != nil {
				log.Printf("Failed to fetch flags after update: %v", err)
			}
		}
	}
}

func (c *Client) BoolVariation(key string, context engine.UserContext, defaultValue bool) bool {
	res := c.evaluate(key, context)
	if res == nil || res.Value == nil {
		return defaultValue
	}
	var val bool
	if err := json.Unmarshal(res.Value, &val); err != nil {
		return defaultValue
	}
	return val
}

func (c *Client) StringVariation(key string, context engine.UserContext, defaultValue string) string {
	res := c.evaluate(key, context)
	if res == nil || res.Value == nil {
		return defaultValue
	}
	var val string
	if err := json.Unmarshal(res.Value, &val); err != nil {
		// Try unmarshaling as raw string if it's just "value"
		return string(bytes.Trim(res.Value, "\""))
	}
	return val
}

func (c *Client) IntVariation(key string, context engine.UserContext, defaultValue int) int {
	res := c.evaluate(key, context)
	if res == nil || res.Value == nil {
		return defaultValue
	}
	var val int
	if err := json.Unmarshal(res.Value, &val); err != nil {
		return defaultValue
	}
	return val
}

func (c *Client) JSONVariation(key string, context engine.UserContext, defaultValue json.RawMessage) json.RawMessage {
	res := c.evaluate(key, context)
	if res == nil || res.Value == nil {
		return defaultValue
	}
	return res.Value
}

func (c *Client) evaluate(key string, context engine.UserContext) *engine.EvaluationResult {
	c.mu.RLock()
	config, ok := c.flags[key]
	c.mu.RUnlock()

	if !ok {
		return nil
	}

	result, err := c.evaluator.Evaluate(config, context)
	if err != nil {
		log.Printf("Evaluation error for flag %s: %v", key, err)
		return nil
	}

	return result
}
