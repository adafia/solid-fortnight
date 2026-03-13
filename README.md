# Solid Fortnight

Solid Fortnight is a feature flagging system designed to provide dynamic control over application features. It allows developers to roll out new features to a subset of users, perform A/B testing, and quickly toggle features on or off without deploying new code.

## Getting Started

## Project Structure

The project is organized into several key directories:

- **`apps/`**: Contains individual microservices for different functionalities (e.g., `analytics`, `evaluator`, `gateway`, `management`, `streamer`).
- **`cmd/`**: Houses the main executable commands, such as the `dashboard`.
- **`deployments/`**: Stores configuration files for deployment, including `config.yaml` and `docker-compose.yml`.
- **`docs/`**: Documentation files.
- **`internal/`**: Internal libraries and packages used across the project, including `config`, `engine`, `protocol`, and `storage`.
- **`scripts/`**: Various utility scripts for tasks like creating, deleting, getting, and updating flags.
- **`sdk/`**: Software Development Kits for different languages/platforms (e.g., `client-js`, `server-go`, `server-python`).

## Running the Application

More detailed instructions for setting up and running the application will be provided here. This typically involves:

1. **Prerequisites**: Go, Docker, etc.
2. **Building**: Instructions to build the various services.
3. **Configuration**: How to set up necessary environment variables or configuration files.
4. **Running Services**: Commands to start the individual `apps` services and the `dashboard`.
