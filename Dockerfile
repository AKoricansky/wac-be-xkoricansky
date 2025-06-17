FROM alpine:3.18 AS builder

# Install build dependencies
RUN apk add --no-cache curl tar gcc musl-dev

# Install Go 1.24.3
RUN mkdir -p /usr/local/go \
    && curl -sSL https://go.dev/dl/go1.24.3.linux-amd64.tar.gz | tar -C /usr/local -xzf -

# Set Go environment variables
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/go"
ENV PATH="${GOPATH}/bin:${PATH}"

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/ambulance-counseling-api-service ./cmd/ambulance-counseling-api-service/main.go

# Use a minimal alpine image for the final stage
FROM alpine:3.18

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/ambulance-counseling-api-service .

# Expose the API port
EXPOSE ${AMBULANCE_COUNSELING_API_PORT}

# Run the API service
CMD ["/app/ambulance-counseling-api-service"]