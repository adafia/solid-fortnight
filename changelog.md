# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Implemented Streamer service in `apps/streamer` for real-time flag updates via Server-Sent Events (SSE).
- Integrated Redis for Pub/Sub messaging and shared configuration.
- Created `internal/storage/pubsub` package for publishing environment updates from the Management API.
- Added Redis connection configuration to `internal/config` and `deployments/config.yaml`.
- Added Dockerfiles for `streamer` and `evaluator` services.
- Updated `docker-compose.yml` to include Redis and the new services.
- Added `run-streamer` and `run-evaluator` commands to the `Makefile`.
- Created a test utility script `scripts/test_sse.go` for verifying real-time flag update streams.
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

- Updated `FlagsHandler` in Management API to publish environment updates on flag configuration changes.
- Enhanced `Makefile` to include Redis and the new services in `start-db` and `stop-db`.
- Separated project and environment management logic in `apps/management/handlers/projects.go`.
- Updated `apps/management/main.go` to route environment requests to the new `EnvironmentsHandler`.
- Updated `apps/management/handlers/environments_test.go` to use the new handler.
- Refactored `apps/management/handlers/flags_test.go` to support testing multiple handlers.
- Standardized Bruno collection with consistent variable naming (`base_url`, `evaluator_base_url`) and logical sequencing.
- Updated Bruno environment configurations for Local and Docker environments.

### Fixed

- Resolved a database insertion failure in environment tests caused by an invalid UUID in the `CreatedBy` field.
