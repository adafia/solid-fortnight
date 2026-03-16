# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy the entire workspace to handle go.work and local module replacements
COPY . .

# Download dependencies
RUN go mod download

# Build the analytics service
RUN cd apps/analytics && go build -o analytics-service main.go

# Run stage
FROM alpine:latest

WORKDIR /app

# Copy binary and config
COPY --from=builder /app/apps/analytics/analytics-service .
COPY --from=builder /app/deployments/config.yaml ./deployments/config.yaml

# Set environment variables
ENV CONFIG_PATH=/app/deployments/config.yaml

# Expose the analytics service port
EXPOSE 8081

CMD ["./analytics-service"]
