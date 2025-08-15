# Build stage
FROM golang:1.24-alpine AS builder

# Install git for go modules
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build argument for version tag
ARG TAG=development

# Build the binary with version tag
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-X 'main.Version=$TAG'" -o slash-prompt .

# Final stage
FROM alpine:3.22.1

# Link to the source code repository
LABEL org.opencontainers.image.source=https://github.com/appgardenstudios/slash-prompt

# Install ca-certificates for HTTPS requests and git for repository cloning
RUN apk --no-cache add ca-certificates git openssh-client

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/slash-prompt .

# Change ownership to non-root user
RUN chown appuser:appgroup /app/slash-prompt

# Switch to non-root user
USER appuser

# Set entrypoint
ENTRYPOINT ["./slash-prompt"]