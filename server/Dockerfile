# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies for SQLite
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod ./
RUN go mod tidy

# Copy source code
COPY . .

# Build the main server
RUN CGO_ENABLED=1 GOOS=linux go build -o server main.go

# Build the cmd tools
RUN CGO_ENABLED=1 GOOS=linux go build -o init-tools ./cmd/init-tools
RUN CGO_ENABLED=1 GOOS=linux go build -o seed ./cmd/seed
RUN CGO_ENABLED=1 GOOS=linux go build -o sync ./cmd/sync

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies for HTTPS requests and changelog fetching
RUN apk --no-cache add ca-certificates curl tzdata

# Copy binaries from builder
COPY --from=builder /app/server .
COPY --from=builder /app/init-tools .
COPY --from=builder /app/seed .
COPY --from=builder /app/sync .

# Create data directory
RUN mkdir -p /data

# Set environment variables
ENV DATA_DIR=/data
ENV PORT=8080

EXPOSE 8080

CMD ["./server"]
