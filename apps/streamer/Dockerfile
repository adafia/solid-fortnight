# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy the entire workspace to handle go.work and local module replacements
COPY . .

# Download dependencies
RUN go mod download

# Build the streamer service
RUN go build -o streamer-service ./apps/streamer/main.go

# Run stage
FROM alpine:latest

WORKDIR /app

# Copy binary and config
COPY --from=builder /app/streamer-service .
COPY --from=builder /app/deployments/config.yaml ./deployments/config.yaml

# Set environment variables for the application to find its assets
ENV CONFIG_PATH=/app/deployments/config.yaml

# Expose the streamer service port
EXPOSE 8084

CMD ["./streamer-service"]
