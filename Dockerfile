FROM golang:latest AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/ambulance-counseling-api-service ./cmd/ambulance-counseling-api-service/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/ambulance-counseling-api-service .

# Expose the API port (hardcoded for build, actual port is used in runtime)
EXPOSE 8080

# Run the API service
CMD ["/app/ambulance-counseling-api-service"]