# Solid Fortnight [WIP]

> **⚠️ Warning:** This project is in its early stages of development and is **not ready for use**. It is a work in progress; expect frequent breaking changes and incomplete features.

*Name suggested by GitHub during repository creation.*

Solid Fortnight is a feature flagging system designed to provide dynamic control over application features. It allows developers to roll out new features to a subset of users, perform A/B testing, and quickly toggle features on or off without deploying new code.

## Currently Implemented Features

### 🛠️ Management API
A RESTful API for administrative operations, including:
- **Project Management**: Create, list, and retrieve projects.
- **Environment Management**: Define multiple environments (e.g., Development, Staging, Production) per project.
- **Flag Management**: Full CRUD operations for feature flags, including:
  - Boolean and Multivariate flag support.
  - Environment-specific flag configurations and overrides.
  - Tagging and metadata support.

### 🏗️ Infrastructure & Core
- **Database**: PostgreSQL integration with automated migrations using `golang-migrate`.
- **Configuration**: Dynamic configuration management via YAML and environment variables.
- **Local Development**: Comprehensive `Makefile` and Docker Compose setup for quick start.
- **API Documentation & Testing**: Ready-to-use **Bruno** collection for exploring and testing the Management API.

### 🔌 SDKs (Work in Progress)
- **Go Server SDK**: Initial structure for server-side evaluation.

## Project Structure

To get started with Solid Fortnight, follow these steps:

1. **Prerequisites**: Ensure you have Go (1.25.0+), Docker, and Docker Compose installed.
2. **Environment Setup**:
   - Copy the example environment file: `cp .env.example .env`
   - Edit `.env` to set your local database credentials.
   - Load the variables into your current shell: `source ./scripts/load_env.sh`
3. **Database**: Start the PostgreSQL database:
   ```bash
   make start-db
   ```
4. **Run the Application**: Start the management service:
   ```bash
   make run-app
   ```

## Project Structure

The project is organized into several key directories:

- **`apps/`**: Contains individual microservices (e.g., `management`).
- **`deployments/`**: Stores configuration files like `config.yaml` and `docker-compose.yml`.
- **`internal/`**: Shared internal libraries (config, storage, etc.).
- **`scripts/`**: Utility scripts for environment management and API testing.

## Running the Application

### 1. Database
The application requires a PostgreSQL database. The provided `docker-compose.yml` sets up a database named `solid_fortnight`.

```bash
make start-db
```

### 2. Configuration
The application uses `deployments/config.yaml` for configuration. Environment variables in this file (like `${DB_USER}`) are expanded at runtime using the values from your `.env` file.

### 3. Management Service
The management service handles flag creation and management.

**Run locally (Go):**
```bash
make run-app
```

**Run with Docker Compose:**
```bash
make start-all
```

### 4. API Tests
This project uses **Bruno** for API testing. The collection is located in the `bruno/` directory.

1.  Open **Bruno**.
2.  Click **Open Collection** and select the `bruno/` folder.
3.  Select the **Local** environment from the environment dropdown.
4.  Use the **Create Project** request to create your first project.
5.  Copy the `id` from the response and use it in the **Create Flag** request.
