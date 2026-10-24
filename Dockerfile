# ----------------------------------------------------
# Stage 1: Build stage
# ----------------------------------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and build dependencies if necessary
RUN apk add --no-cache git

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Compile static binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/mail-api ./cmd/api

# ----------------------------------------------------
# Stage 2: Runtime stage
# ----------------------------------------------------
FROM alpine:latest

# Install ca-certificates (required for STARTTLS / SSL certificate verification)
# and tzdata (for timezone support)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Create a non-privileged user and group for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy the compiled binary from the builder stage
COPY --from=builder /app/mail-api /app/mail-api

# Change ownership to the non-root user
USER appuser

# Expose default HTTP port
EXPOSE 8082

# Run the microservice
ENTRYPOINT ["/app/mail-api"]
