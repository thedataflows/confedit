# Build stage
FROM goreleaser/goreleaser:v2.12.5 AS builder

# Set working directory
WORKDIR /app

# Copy source code and goreleaser config
COPY . .

# Build the application using goreleaser
# Use --snapshot for local builds without git tags
# Filter for linux/amd64 only to speed up Docker build
ARG VERSION=v0.0.0-snapshot
ENV GORELEASER_CURRENT_TAG=$VERSION
RUN goreleaser build --snapshot --clean --single-target

# Final stage
FROM alpine:3.22

# Install ca-certificates for HTTPS requests
RUN apk add --no-cache ca-certificates

# Create non-root user
RUN addgroup -g 1000 confedit && \
    adduser -D -s /bin/sh -u 1000 -G confedit confedit

# Create directories for data and snapshots
RUN mkdir -p /data /data/snapshots && \
    chown -R confedit:confedit /data

# Copy binary from builder stage
COPY --from=builder /app/dist/default_linux_amd64_v1/confedit /usr/local/bin/confedit

# Set user
USER confedit

# Set working directory
WORKDIR /data

# Default command
ENTRYPOINT ["confedit"]
CMD ["--help"]
