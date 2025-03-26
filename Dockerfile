FROM golang:1.21-alpine as builder

WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/btcpay-service ./cmd/service

# Use alpine for smaller final image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/btcpay-service .
# Copy config files
COPY --from=builder /app/config ./config

EXPOSE 8080

# Run the application
CMD ["/app/btcpay-service"]