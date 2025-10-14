# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations for size and performance
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -a -installsuffix cgo \
    -o k8s-deployment-exporter .

# Final stage - minimal runtime image
FROM alpine:3.18

# Add ca-certificates and timezone data for HTTPS requests and WIB timezone
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/k8s-deployment-exporter .

# Run as non-root user
RUN addgroup -g 1000 exporter && \
    adduser -D -u 1000 -G exporter exporter && \
    chown -R exporter:exporter /root

USER exporter

EXPOSE 9101

ENTRYPOINT ["./k8s-deployment-exporter"]
