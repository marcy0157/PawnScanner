
# Stage 1: Build the Go application
FROM golang:1.23.2 AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o main ./cmd/main.go

# Stage 2: Build the final image
FROM debian:bookworm-slim

# Set the working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/main .

# Copy the static web files
COPY web ./web

# Expose the port the app runs on
EXPOSE 8080

# Command to run the application
CMD ["./main"]
