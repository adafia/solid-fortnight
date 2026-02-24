# Directory Structure

```plaintext
/solid-fortnight
├── go.work                # Orchestrates all modules
├── apps/
│   ├── gateway/           # Gin/Echo - Auth & Rate Limiting
│   ├── management/        # Gorm - CRUD for Flags/Rules
│   ├── evaluator/         # High-perf Rule Engine (GRPC/Internal)
│   ├── streamer/          # Gorilla WebSocket / SSE
│   └── analytics/         # Worker to process MQ -> TimescaleDB
├── cmd/
│   └── dashboard/         # React/Vite Frontend
├── sdk/
│   ├── client-js/         # (Stay in JS/TS)
│   ├── server-go/         # Go SDK for other internal apps
│   └── server-python/     # Python SDK (FFI or RPC)
├── internal/              # Code that cannot be imported externally
│   ├── protocol/          # Protobuf / Shared Event Definitions
│   ├── engine/            # The actual logic that calculates toggles
│   └── storage/           # Shared DB drivers (Postgres/Redis)
├── deployments/           # Dockerfiles & K8s manifests
└── scripts/               # Migration and Seed scripts
```
