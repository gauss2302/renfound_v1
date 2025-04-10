FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o renfound_app ./cmd/main.go

# Use a minimal alpine image for the final container
FROM alpine:3.19

# Add security and runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl \
    && addgroup -S appgroup \
    && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/renfound_app .

# Copy migration files
COPY migrations ./migrations

# Install golang-migrate
RUN apk add --no-cache curl \
    && curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz \
    && mv migrate /usr/local/bin/migrate \
    && chmod +x /usr/local/bin/migrate

# Create a non-root user
USER appuser

# Expose the application port
EXPOSE 8090

# Command to run the application
CMD ["./renfound_app"]