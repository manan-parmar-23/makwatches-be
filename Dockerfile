# Builder stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install necessary tools
RUN apk add --no-cache git make

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pehnaw-be ./cmd/api

# Final stage
FROM alpine:3.18

# Set working directory
WORKDIR /app

# Install necessary runtime packages
RUN apk --no-cache add ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/pehnaw-be .

# Copy example.env file
COPY example.env .env

# Expose port
EXPOSE 8080

# Command to run the application
CMD ["./pehnaw-be"]