# Go Server SDK

The Go Server SDK provides a high-performance way for Go applications to evaluate feature flags locally with real-time updates.

## Features

- **Local Evaluation**: Sub-millisecond flag evaluation using an in-memory cache.
- **Real-time Updates**: Immediate configuration sync via Server-Sent Events (SSE).
- **Graceful Fallback**: Periodic polling ensures the SDK eventually recovers if SSE connection is lost.
- **Support for All Flag Types**: Boolean, String, Integer, and JSON variations.

## Installation

```bash
go get github.com/adafia/solid-fortnight/sdk/server-go
```

## Usage

### 1. Initialize the Client

```go
import (
    "github.com/adafia/solid-fortnight/sdk/server-go"
    "github.com/adafia/solid-fortnight/internal/engine"
)

client, err := sdk.NewClient(sdk.Config{
    EvaluatorURL:  "http://localhost:8082",
    StreamerURL:   "http://localhost:8084",
    EnvironmentID: "your-environment-uuid",
})
defer client.Close()
```

### 2. Evaluate a Flag

```go
userContext := engine.UserContext{
    ID: "user-123",
    Attributes: map[string]interface{}{
        "email": "user@example.com",
        "plan":  "premium",
    },
}

// Boolean flag
if client.BoolVariation("new-feature", userContext, false) {
    // Feature is enabled
}

// String flag
theme := client.StringVariation("app-theme", userContext, "light")

// Integer flag
maxItems := client.IntVariation("max-items", userContext, 10)
```

## Configuration Options

| Option | Type | Description | Default |
| :--- | :--- | :--- | :--- |
| `EvaluatorURL` | `string` | URL of the Evaluator service. | Required |
| `StreamerURL` | `string` | URL of the Streamer service. | Required |
| `EnvironmentID`| `string` | UUID of the environment. | Required |
| `PollInterval` | `time.Duration` | Fallback polling interval. | `5 minutes` |
