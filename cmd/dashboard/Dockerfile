# Build stage
FROM oven/bun:1.1 AS builder

WORKDIR /app

# Copy package files and install dependencies
COPY package.json bun.lock ./
RUN bun install

# Copy source and build
COPY . .
RUN bun run build

# Run stage
FROM caddy:2-alpine

# Copy build output to caddy's default root
COPY --from=builder /app/dist /usr/share/caddy

# Copy Caddyfile
COPY Caddyfile /etc/caddy/Caddyfile

EXPOSE 80

CMD ["caddy", "run", "--config", "/etc/caddy/Caddyfile"]
