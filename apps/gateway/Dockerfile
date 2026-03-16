# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy the entire workspace to handle go.work and local module replacements
COPY . .

# Download dependencies
RUN go mod download

# Build the gateway service
RUN go build -o gateway-service ./apps/gateway/main.go

# Run stage
FROM alpine:latest

WORKDIR /app

# Install config
COPY --from=builder /app/gateway-service .
COPY --from=builder /app/deployments/config.yaml ./deployments/config.yaml

# Set environment variables for the application to find its assets
ENV CONFIG_PATH=/app/deployments/config.yaml

# Expose the gateway service port
EXPOSE 8080

CMD ["./gateway-service"]
