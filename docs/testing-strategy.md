# Testing Strategy

This document outlines the testing strategy for the `solid-fortnight` project, ensuring reliability across the gateway, management, evaluator, streamer, and analytics services, as well as the admin dashboard.

## 1. The Testing Pyramid

We follow a layered testing approach to balance speed and confidence:

| Layer | Tool | Scope | Speed |
| :--- | :--- | :--- | :--- |
| **Unit Tests (Backend)** | Go `testing` | Engine logic, config, utility functions. | Fast (< 1s) |
| **Unit Tests (UI)** | Vitest + RTL | Dashboard components, modal logic, routing. | Fast (< 2s) |
| **Integration Tests** | Go `testing` + Docker | API handlers interacting with real DB/Redis. | Medium (~5s) |
| **End-to-End (E2E)** | Playwright + Docker | Full user flows through Gateway & UI. | Slow (> 30s) |

## 2. Layer Details

### UI Component Testing (Vitest)

Used for the Admin Dashboard to verify UI behavior without a backend.

- **Mocking**: Global `fetch` is mocked to simulate API Gateway responses.
- **Location**: `cmd/dashboard/src/**/*.test.tsx`.
- **Command**: `cd cmd/dashboard && bun run test`.

### Integration Testing (Synchronous)

Used for the Management API and Gateway. These tests ensure that HTTP requests correctly modify the PostgreSQL state or proxy correctly to internal services.

- **Infrastructure**: Real PostgreSQL/Redis containers via `make test-db-up`.
- **Location**: `apps/*/handlers/*_test.go`.

### Browser-based E2E Testing (Playwright)

The final verification layer that tests the real integrated system.

- **Environment**: Full Docker Compose stack (`make start-all`).
- **Location**: `cmd/dashboard/tests/e2e/*.spec.ts`.
- **Command**: `make test-e2e`.

## 3. Mocking Strategy

- **Backend**: Prefer real infrastructure (Postgres/Redis) for Go tests.
- **Gateway**: Use `httptest.NewServer` to mock internal services when testing proxy logic.
- **UI**: Mock `fetch` in Vitest to test error states and complex data scenarios without starting Go services.

## 4. Execution Guide

### Fast Feedback Loop (Recommended for Development)

```bash
# 1. Test Backend Logic
go test ./internal/engine/... ./internal/config/...

# 2. Test UI Logic
cd cmd/dashboard && bun run test
```

### Full Verification (Before Commit)

```bash
# 1. Run all Backend Integration Tests
make test

# 2. Run Full Stack E2E
make test-e2e
```
