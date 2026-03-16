# Solid Fortnight - Requirements Document

---

## Project Overview

A production-ready feature flag service that enables development teams to safely deploy features, conduct A/B tests, and control feature rollouts across distributed systems.

> This project demonstrates advanced backend engineering skills including distributed systems design, real-time data synchronization, API design, caching strategies, and operational excellence.

### Core Objectives

- Build a horizontally scalable feature flag service that can handle 10,000+ requests per second
- Implement sophisticated targeting rules (user segments, percentage rollouts, geographic targeting, custom attributes)
- Design a low-latency evaluation engine with sub-5ms p99 response times
- Create a comprehensive management API and admin dashboard
- Demonstrate operational excellence with monitoring, alerting, and disaster recovery

---

## Functional Requirements

### 1. Flag Management

#### Flag CRUD Operations

- Create, read, update, and delete feature flags with version history
- Support multiple flag types: boolean, string, number, JSON
- Flag metadata: name, description, tags, owner, creation/modification timestamps
- Soft delete with archival and restore capabilities

#### Environment Management

- Support multiple environments (development, staging, production) with independent configurations
- Environment-specific default values and targeting rules
- Promote flag configurations between environments

---

### 2. Targeting Rules Engine

#### User Targeting

- Individual user targeting by user ID or email
- User segment targeting (e.g., premium users, beta testers)
- Custom attribute matching (user properties like plan_type, signup_date, region)

#### Rollout Strategies

- Percentage-based rollouts with consistent hashing (same user always gets same result)
- Gradual rollout schedules (auto-increment percentage over time)
- Geographic targeting by country, region, or city
- Time-based activation (enable/disable on schedule)

#### Rule Composition

- Combine multiple rules with AND/OR logic
- Rule prioritization (first-match-wins evaluation)
- Fallback values when no rules match

---

### 3. Flag Evaluation API

#### Evaluation Endpoints

- Single flag evaluation: `POST /api/v1/evaluate` (returns variation for a context) [IMPLEMENTED]
- Bulk evaluation: `POST /api/v1/flags/evaluate` (returns all flags for a context) [PLANNED]
- Server-side SDKs can cache and evaluate locally after initial sync

#### Evaluation Context

- Accept user context: userId, email, attributes (key-value pairs)
- Accept application context: environment, version, region
- Return evaluated value, variation key, and reason for decision

---

### 4. Analytics & Monitoring

- Track flag evaluation events (which users see which variations)
- Aggregate metrics: evaluation counts, unique users per variation, error rates
- Flag usage dashboard showing active vs. stale flags
- Audit log of all flag configuration changes

---

### 5. SDK Requirements

- Server-side SDK (Go and/or Node.js) with local caching and real-time updates
- Client-side SDK (JavaScript) for browser applications
- Automatic reconnection and error handling
- Graceful degradation when service is unavailable (use last known values)

---

### 6. Admin Dashboard

- Web UI for managing flags, environments, and targeting rules
- Visual rule builder (drag-and-drop or form-based)
- Real-time preview of affected users before saving changes
- Search, filter, and sort flags by tags, owner, or status

---

## Technical Requirements

### Performance

- Flag evaluation latency: p50 < 2ms, p99 < 5ms
- Support 10,000+ requests per second per instance
- Multi-layer caching: in-memory, Redis, CDN (for client-side SDKs)

### Scalability

- Horizontally scalable stateless API servers
- Database sharding strategy for high write volumes (analytics events)
- Async event processing with message queue for analytics ingestion

### Availability & Reliability

- 99.9% uptime SLA
- Multi-region deployment with automatic failover
- Circuit breakers for external dependencies
- Graceful degradation: SDKs use cached values if service unavailable

### Security

- API authentication via API keys and JWT tokens
- Role-based access control (RBAC) for dashboard and management API
- Rate limiting to prevent abuse
- Audit logging of all administrative actions

### Data Consistency

- Eventual consistency for flag configuration propagation (acceptable delay: <100ms) [IMPLEMENTED]
- Strong consistency for flag CRUD operations [IMPLEMENTED]
- Real-time updates via Server-Sent Events (SSE) [IMPLEMENTED]

---

## System Architecture

### Core Components

| Component | Status | Description |
| ----------- | ------- | ------------- |
| **API Gateway** | [PLANNED] | Entry point for all requests. Handles authentication, rate limiting. |
| **Evaluation Service** | [IMPLEMENTED] | Core service that evaluates flags based on targeting rules. |
| **Management Service** | [IMPLEMENTED] | Handles CRUD operations for flags, environments, and rules. |
| **Analytics Service** | [PLANNED] | Ingests evaluation events, computes metrics, powers dashboard. |
| **Stream Service** | [IMPLEMENTED] | Manages SSE connections for real-time flag updates to SDKs. |
| **Admin Dashboard** | [PLANNED] | React-based web UI for managing flags and viewing analytics. |

### Data Storage

| Store | Purpose | Status |
| ------- | --------- | ------- |
| **PostgreSQL** | Primary database for configurations, targeting rules, logs. | [IMPLEMENTED] |
| **Redis** | Pub/Sub messaging and shared configuration. | [IMPLEMENTED] |
| **TimescaleDB** | Time-series database for analytics events and metrics. | [PLANNED] |
| **Message Queue** | Redis for change notifications and async processing. | [IMPLEMENTED] |

### Technology Stack (Recommended)

- **Backend**: Go (for ultra-low latency)
- **Database**: PostgreSQL (primary), Redis (cache), TimescaleDB (analytics)
- **Message Queue**: Redis Streams, NATS, or RabbitMQ
- **Admin Dashboard**: React with TypeScript, TailwindCSS
- **SDKs**: Go and JavaScript (can add more later)
- **Containerization**: Docker + Docker Compose for local dev, Kubernetes for production
- **Observability**: Prometheus + Grafana for metrics, ELK stack or Loki for logs
