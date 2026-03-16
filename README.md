# Solid Fortnight [WIP]

> **⚠️ Warning:** This project is in its early stages of development and is **not ready for use**. It is a work in progress; expect frequent breaking changes and incomplete features.

*Name suggested by GitHub during repository creation.*

Solid Fortnight is a feature flagging system designed to provide dynamic control over application features. It allows developers to roll out new features to a subset of users, perform A/B testing, and quickly toggle features on or off without deploying new code.

## Currently Implemented Features

### 🛠️ API Gateway Service

A high-performance reverse proxy that:

- Serves as the single entry point (port 8080) for all external requests.
- Handles dynamic **Service Discovery** for Management, Evaluator, Streamer, and Analytics services.
- Implements **Middleware Chaining** for request logging, authentication (API Key/JWT), and rate limiting.
- Manages **Path Mapping** to simplify SDK and Admin UI integrations.

### 🛠️ Management API

A RESTful API for administrative operations, including:

- **Project Management**: Create, list, and retrieve projects.
- **Environment Management**: Define multiple environments (e.g., Development, Staging, Production) per project.
- **Flag Management**: Full CRUD operations for feature flags, including multivariate support and environment-specific overrides.

### 🎯 Evaluator Service

A high-performance evaluation engine that:

- Supports multi-clause targeting rules (EQUALS, IN, CONTAINS, etc.).
- Implements consistent percentage-based rollouts using MD5 hashing.
- Provides sub-millisecond evaluation for single flag requests.

### 📡 Streamer Service

A real-time synchronization service that:

- Uses **Server-Sent Events (SSE)** to push flag updates to SDKs instantly.
- Powered by a **Redis Pub/Sub** backbone for cross-service communication.
- Maintains persistent, high-concurrency client connections with heartbeats.

### 📊 Analytics Service

A high-throughput event ingestion and processing service that:

- Buffers evaluation events using **Redis Streams** for sub-millisecond ingestion.
- Processes events asynchronously via a **Background Worker** with consumer groups.
- Persists events in **PostgreSQL** for long-term A/B testing analysis and metrics.
- Supports bulk ingestion for reduced network overhead.

### 🏗️ Infrastructure & Core

- **Database**: PostgreSQL for persistent storage and **Redis** for real-time messaging.
- **Automated Migrations**: Database schema management using `golang-migrate`.
- **Configuration**: Dynamic configuration via YAML and environment variables.
- **Local Development**: Comprehensive `Makefile` and Docker Compose setup for quick start.
- **API Documentation**: Ready-to-use **Bruno** collection for exploring and testing the Management, Evaluator, Streamer, and Analytics APIs.

### 🔌 SDKs (Work in Progress)

- **Go Server SDK**: Initial structure for server-side evaluation with local caching and real-time SSE updates.

## Project Setup

To get started with Solid Fortnight, follow these steps:

1. **Prerequisites**: Ensure you have Go (1.25.0+), Docker, and Docker Compose installed.
2. **Environment Setup**:
   - Copy the example environment file: `cp .env.example .env`
   - Edit `.env` to set your local database and Redis credentials.
   - Load the variables into your current shell: `source ./scripts/load_env.sh`
3. **Infrastructure**: Start the PostgreSQL and Redis databases:

   ```bash
   make start-db
   ```

4. **Run the Application**: You can run the services individually or all together:
   - **Management**: `make run-app`
   - **Evaluator**: `make run-evaluator`
   - **Streamer**: `make run-streamer`
   - **Analytics**: `make run-analytics`
   - **Gateway**: `make run-gateway`
   - **All (Docker)**: `make start-all`
## Project Structure

The project is organized into several key directories:

- **`apps/`**: Individual microservices (`gateway`, `management`, `evaluator`, `streamer`, `analytics`).
- **`deployments/`**: Stores configuration files like `config.yaml` and `docker-compose.yml`.
- **`internal/`**: Shared internal libraries (config, evaluation engine, storage drivers, pubsub, protocol).
- **`docs/`**: Detailed service documentation and testing strategies.

## Running the Application

### 1. Infrastructure

The application requires PostgreSQL and Redis. The provided `docker-compose.yml` sets up both.

```bash
make start-db
```

### 2. Configuration

The application uses `deployments/config.yaml` for configuration. Environment variables in this file (like `${DB_USER}`) are expanded at runtime using the values from your `.env` file.

### 3. API Tests

This project uses **Bruno** for API testing. The collection is located in the `bruno/` directory.

1. Open **Bruno**.
2. Click **Open Collection** and select the `bruno/` folder.
3. Select the **Local** environment from the environment dropdown.
4. Use the **Management API** to create projects and flags.
5. Use the **Evaluator API** to test targeting rules.
6. Use the **Streamer API** to watch real-time updates.
7. Use the **Analytics API** to send evaluation events.
