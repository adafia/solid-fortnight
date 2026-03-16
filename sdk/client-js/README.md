# Solid Fortnight JS SDK

The Solid Fortnight JS SDK provides high-performance feature flag evaluation for JavaScript and TypeScript applications, powered by Bun.

## Features

- **Local Evaluation**: Sub-millisecond evaluation with in-memory caching.
- **Real-time Updates**: Instant synchronization using Server-Sent Events (SSE).
- **Universal**: Works in Node.js, Bun, and browsers (with appropriate polyfills if needed).
- **Type-safe**: Built with TypeScript for excellent developer experience.

## Installation

```bash
bun add github.com/adafia/solid-fortnight/sdk/client-js
# or
npm install github.com/adafia/solid-fortnight/sdk/client-js
```

## Usage

### 1. Initialize the Client

```typescript
import { Client } from "@solid-fortnight/client-js";

const client = new Client({
  evaluatorUrl: "http://localhost:8082",
  streamerUrl: "http://localhost:8084",
  environmentId: "your-environment-uuid",
});

await client.init();
```

### 2. Evaluate Flags

```typescript
const context = {
  id: "user-123",
  attributes: {
    email: "user@example.com",
    plan: "premium",
  },
};

// Boolean flag
const isEnabled = client.boolVariation("new-feature", context, false);

// String flag
const theme = client.stringVariation("app-theme", context, "light");

// Number flag
const maxItems = client.numberVariation("max-items", context, 10);

// JSON flag
const config = client.jsonVariation("complex-config", context, { default: true });
```

### 3. Cleanup

```typescript
client.close();
```

## Configuration

| Option | Type | Description | Default |
| :--- | :--- | :--- | :--- |
| `evaluatorUrl` | `string` | URL of the Evaluator service. | Required |
| `streamerUrl` | `string` | URL of the Streamer service. | Required |
| `environmentId`| `string` | UUID of the environment. | Required |
| `pollInterval` | `number` | Fallback polling interval in ms. | `300000` (5 mins) |

## Development

```bash
bun install
bun test
```
