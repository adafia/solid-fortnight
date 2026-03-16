# Streamer Service

The Streamer service is a high-performance, real-time synchronization component of the Solid Fortnight feature flagging system. It enables SDKs and client applications to receive immediate updates when feature flag configurations change, without polling.

## Architecture

The service uses **Server-Sent Events (SSE)** to push updates to connected clients. It acts as a bridge between the **Management API** (the producer) and the **SDKs** (the consumers), using **Redis Pub/Sub** as the messaging backbone.

### Flow of Events

1. **Change Trigger**: A user updates a flag configuration or variation via the Management API.
2. **Publish**: The Management API publishes an `environment_update` message to Redis on the `environment_updates` channel. This message includes the updated flag configuration.
3. **Subscribe**: The Streamer service is subscribed to this Redis channel.
4. **Broadcast**: Upon receiving the message, the Streamer identifies all clients currently connected to that specific `environment_id` and broadcasts the JSON payload directly via their SSE connection.

## API Specification

### GET `/stream`

Establish a persistent SSE connection to receive updates for a specific environment.

**Query Parameters:**

| Parameter | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `environment_id` | UUID | Yes | The ID of the environment to subscribe to. |

**Response Headers:**

- `Content-Type: text/event-stream`
- `Cache-Control: no-cache`
- `Connection: keep-alive`

**Message Format:**

The service sends JSON messages prefixed with `data:`.

- **Delta Update Event**:

  ```json
  data: {"environment_id": "uuid", "data": {"key": "flag-key", "enabled": true, ...}}\n\n
  ```
  
- **Generic Update Event (Fallback)**: `data: update\n\n`
- **Keep-alive (Heartbeat)**: `: keep-alive\n\n` (Sent every 30 seconds to maintain the connection).

## Implementation Details

- **Language**: Go
- **Concurrency**: Each client connection is handled in its own goroutine. The service uses a central `Hub` with `sync.RWMutex` to manage client registrations and broadcasts safely.
- **Backpressure**: The service uses buffered channels (size 10) for each client. If a client is too slow to receive messages, the broadcast to that specific client is skipped to prevent blocking the entire broadcast loop.
- **Fault Tolerance**: If the Redis connection is lost, the service attempts to reconnect every 2 seconds.

## Configuration

The service is configured via `deployments/config.yaml` and environment variables:

| Environment Variable | Description | Default |
| :--- | :--- | :--- |
| `REDIS_ADDR` | Address of the Redis server | `localhost:6379` |
| `REDIS_PASSWORD` | Password for Redis | `""` |
| `CONFIG_PATH` | Path to the YAML configuration file | `deployments/config.yaml` |

## Local Development & Testing

### Running the Service

```bash
make run-streamer
```

### Testing with the Test Script

A utility script is provided to simulate an SDK connection:

```bash
go run scripts/test_sse.go <environment_id>
```

### Testing with Bruno

The **Bruno** collection includes a **Streamer API** folder with a **Stream Environment** request.

1. Open the Bruno collection.
2. Select the **Local** or **Docker** environment.
3. Set the `environment_id` variable.
4. Send the **Stream Environment** request.
5. Bruno will keep the connection open and display real-time events as they arrive.

### Manual Verification

You can also test using `curl`:

```bash
curl -N "http://localhost:8084/stream?environment_id=<your-env-id>"
```
