# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Implemented **API Gateway** (`apps/gateway`) to serve as a unified entry point (port 8080) for all services.
- Implemented **Middleware Chain** in the gateway with support for **Logging**, **Authentication (API Key/JWT)**, and **Rate Limiting**.
- Added **Dynamic Service Discovery** to the gateway via environment variables.
- Added **Path Mapping** in the gateway to route external `/api/v1` requests to internal services.
- Added comprehensive **Unit and Integration Tests** for the API Gateway.
- Added **Analytics API Documentation** to the Bruno collection.
- Updated **Bruno Collection** to route all requests through the API Gateway by default.
- Implemented **Delta Updates via SSE** in the Streamer service and SDKs (`client-js`, `server-go`) to update local cache selectively without refetching all flags on every change.
- Implemented **Analytics Background Worker** (`apps/analytics/service/worker.go`) to consume evaluation events from Redis Streams.
- Implemented **PostgreSQL Persistence** for evaluation events in the Analytics service using batch insertions.
- Added **Performance Benchmarks** for the core evaluation engine (`internal/engine/engine_test.go`).
- Expanded **Testing Strategy** documentation (`docs/testing-strategy.md`) with asynchronous testing patterns and performance monitoring.
- Added end-to-end integration tests for the Analytics pipeline (API -> Redis -> Worker -> PostgreSQL).
- Implemented **JS SDK** (`sdk/client-js`) using Bun and TypeScript, featuring local evaluation and real-time SSE updates.
- Implemented **Go Server SDK** (`sdk/server-go`) with local caching, real-time synchronization via SSE, and fallback polling.
- Implemented **Analytics Service** (`apps/analytics`) for high-throughput evaluation event ingestion using Redis Streams.
- Created `internal/protocol` package for shared event schemas.
- Added database migrations for `evaluation_events` storage.
- Added integration tests for Analytics and Streamer services.
- Implemented Streamer service in `apps/streamer` for real-time flag updates via Server-Sent Events (SSE).
- Integrated Redis for Pub/Sub messaging and shared configuration.
- Created `internal/storage/pubsub` package for publishing environment updates from the Management API.
- Added Redis connection configuration to `internal/config` and `deployments/config.yaml`.
- Added Dockerfiles for `streamer` and `evaluator` services.
- Updated `docker-compose.yml` to include Redis and the new services.
- Added `run-streamer` and `run-evaluator` commands to the `Makefile`.
- Created a test utility script `scripts/test_sse.go` for verifying real-time flag update streams.
- Added **Streamer API** request to the **Bruno** collection for real-time SSE testing.
- Implemented Evaluator service in `apps/evaluator` for flag evaluation.
- Added database migrations for targeting rules and percentage-based rollouts.
- Implemented targeting rule and rollout support in `internal/storage/store/flag_configs.go`.
- Added `GetFlagByKey` and `GetEnvironmentByKey` methods to the storage layer.
- Implemented core evaluation engine in `internal/engine`.
- Added support for multi-clause targeting rules with various operators (EQUALS, IN, CONTAINS, etc.).
- Added consistent percentage-based rollouts using MD5 hashing.
- Comprehensive unit tests for the evaluation engine.
- Added "Currently Implemented Features" section to `README.md`.
- Created `EnvironmentsHandler` to handle environment-specific requests.
- Added `apps/management/handlers/environments.go`.
- Added `apps/management/handlers/projects_test.go` for isolated project testing.
- Added comprehensive Bruno API collection for Management and Evaluator services.
- Added Bruno documentation for environment and flag variation management.
- Added Evaluator API documentation with sample context attributes.

### Changed

- Modified `apps/management` to publish full flag configurations on updates.
- Modified `apps/streamer` to broadcast the JSON payload instead of a generic `"update"` event.
- Modified SDKs (`sdk/client-js`, `sdk/server-go`) to apply delta updates directly to local cache without an HTTP request.
- Updated `docs/streamer-service.md` to reflect the new delta update JSON payload format.
- Updated `FlagsHandler` in Management API to publish environment updates on flag configuration changes.
- Enhanced `Makefile` to include Redis and the new services in `start-db` and `stop-db`.
- Separated project and environment management logic in `apps/management/handlers/projects.go`.
- Updated `apps/management/main.go` to route environment requests to the new `EnvironmentsHandler`.
- Updated `apps/management/handlers/environments_test.go` to use the new handler.
- Refactored `apps/management/handlers/flags_test.go` to support testing multiple handlers.
- Standardized Bruno collection with consistent variable naming (`base_url`, `evaluator_base_url`) and logical sequencing.
- Updated Bruno environment configurations for Local and Docker environments.

### Fixed

- Resolved an `EventSource` mocking issue in `sdk/client-js` tests by dynamically resolving the implementation, allowing SSE delta updates to be tested properly.
- Resolved a database insertion failure in environment tests caused by an invalid UUID in the `CreatedBy` field.
