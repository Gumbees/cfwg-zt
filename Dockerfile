# Docker file for development testing
FROM golang:1.20-alpine

# Install build and runtime dependencies
RUN apk add --no-cache \
    bash \
    curl \
    make \
    gcc \
    libc-dev

# WireGuard related packages (for testing only)
RUN apk add --no-cache \
    wireguard-tools

# Create app directory
WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN make build

# Create necessary directories
RUN mkdir -p /etc/cfwg-zt /etc/wireguard /var/log/cfwg-zt

# Copy default config
RUN cp config.yaml.example /etc/cfwg-zt/config.yaml

# Create a simple entrypoint script
RUN echo '#!/bin/sh\nexec /app/build/cfwg-zt "$@"' > /entrypoint.sh \
    && chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
CMD ["start"]
