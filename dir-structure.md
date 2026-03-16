# Directory Structure

```plaintext
/solid-fortnight
├── go.work                # Orchestrates all modules
├── apps/
│   ├── gateway/           # (Planned) Auth & Rate Limiting
│   ├── management/        # CRUD for Flags/Rules (Gin/HTTP)
│   ├── evaluator/         # High-perf Rule Engine (HTTP)
│   ├── streamer/          # SSE broadcast service
│   └── analytics/         # (Planned) Ingest metrics -> DB
├── cmd/
│   └── dashboard/         # (Planned) React/Vite Frontend
├── sdk/
│   ├── client-js/         # (WIP) Browser SDK
│   ├── server-go/         # (WIP) Go SDK for server-side evaluation
│   └── server-python/     # (Planned) Python SDK
├── internal/              # Shared internal packages
│   ├── engine/            # Rule evaluation logic
│   ├── storage/           # Shared DB drivers (Postgres/Redis/PubSub)
│   └── config/            # YAML configuration loader
├── deployments/           # Dockerfiles & Compose manifests
├── docs/                  # Detailed service documentation
└── scripts/               # Utility scripts and test tools
```
