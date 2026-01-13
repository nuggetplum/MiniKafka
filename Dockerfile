# --- Stage 1: Build Module ---
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy dependency files first (better caching)
COPY go.mod ./
# If you had a go.sum, you'd copy it here too: COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary
# CGO_ENABLED=0 ensures a static binary (no dependencies on system libraries)
RUN CGO_ENABLED=0 GOOS=linux go build -o /minikafka cmd/server/main.go

# --- Stage 2: Runtime Module ---
FROM alpine:latest

# Set working directory
WORKDIR /root/

# Copy the compiled binary from the builder stage
COPY --from=builder /minikafka .

# Expose the port our app runs on
EXPOSE 8080

# Command to run when the container starts
CMD ["./minikafka"]