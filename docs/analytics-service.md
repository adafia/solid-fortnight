# Analytics Service

The Analytics service is responsible for ingesting, processing, and storing flag evaluation events. It provides the data necessary for usage monitoring, A/B testing analysis, and flag lifecycle management.

## Architecture

The service is designed for high-throughput ingestion using a **Producer-Consumer** pattern with **Redis Streams** as the intermediate buffer.

### Flow of Events

1. **Event Generation**: SDKs or the API Gateway generate evaluation events whenever a flag is evaluated.
2. **Batch Ingestion**: The SDKs send events in batches to the Analytics API (`POST /api/v1/events/bulk`).
3. **Queuing**: The Analytics API pushes these events into a Redis Stream (`evaluation_events_stream`) using Redis pipelines for sub-millisecond queuing latency.
4. **Processing**: An asynchronous background worker (`service/worker.go`) consumes events from the Redis Stream using consumer groups.
5. **Storage**: The worker batch-inserts the processed events into the PostgreSQL `evaluation_events` table for persistent storage.

## API Specification

### POST `/api/v1/events/bulk`

Ingest a batch of evaluation events.

**Request Body:**

```json
[
  {
    "project_id": "uuid",
    "environment_id": "uuid",
    "flag_key": "new-ui-feature",
    "user_id": "user_12345",
    "variation_key": "enabled",
    "value": true,
    "reason": "rule match: beta-users",
    "context": {
      "email": "user@example.com",
      "plan": "premium"
    },
    "evaluated_at": 1710582000
  }
]
```

**Response:**

- `202 Accepted`: Events have been successfully queued for processing.
- `400 Bad Request`: Invalid JSON payload or missing required fields.

## Event Schema (`internal/protocol`)

The common event structure used across the system:

| Field | Type | Description |
| :--- | :--- | :--- |
| `project_id` | `string` | UUID of the project. |
| `environment_id` | `string` | UUID of the environment. |
| `flag_key` | `string` | Unique key of the feature flag. |
| `user_id` | `string` | Unique identifier for the user. |
| `variation_key` | `string` | The key of the variation served (e.g., "control", "treatment"). |
| `value` | `json` | The actual value served to the user. |
| `reason` | `string` | The reason for the evaluation result (e.g., "default", "targeting rule"). |
| `context` | `json` | The user attributes used during evaluation. |
| `evaluated_at` | `int64` | Unix timestamp of the evaluation. |

## Storage Schema

Events are eventually persisted in the `evaluation_events` table in PostgreSQL:

| Column | Type | Constraints |
| :--- | :--- | :--- |
| `id` | `BIGSERIAL` | Primary Key |
| `project_id` | `UUID` | Foreign Key (projects) |
| `environment_id` | `UUID` | Foreign Key (environments) |
| `flag_key` | `VARCHAR` | Index |
| `user_id` | `VARCHAR` | Index |
| `variation_key` | `VARCHAR` | |
| `value` | `JSONB` | |
| `reason` | `VARCHAR` | |
| `context` | `JSONB` | |
| `evaluated_at` | `TIMESTAMP` | Index |

## Local Development

### Running the Service

```bash
make run-analytics
```

### Configuration

The service uses the following environment variables:

| Variable | Description | Default |
| :--- | :--- | :--- |
| `REDIS_ADDR` | Redis connection address. | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password. | `""` |
| `CONFIG_PATH` | Path to YAML config. | `deployments/config.yaml` |
