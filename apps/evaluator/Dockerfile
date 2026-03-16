# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy the entire workspace to handle go.work and local module replacements
COPY . .

# Download dependencies
RUN go mod download

# Build the evaluator service
RUN go build -o evaluator-service ./apps/evaluator/main.go

# Run stage
FROM alpine:latest

WORKDIR /app

# Copy binary and config
COPY --from=builder /app/evaluator-service .
COPY --from=builder /app/deployments/config.yaml ./deployments/config.yaml
COPY --from=builder /app/internal/storage/migrations ./internal/storage/migrations

# Set environment variables for the application to find its assets
ENV CONFIG_PATH=/app/deployments/config.yaml
ENV MIGRATIONS_PATH=/app/internal/storage/migrations

# Expose the evaluator service port
EXPOSE 8082

CMD ["./evaluator-service"]
