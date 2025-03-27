FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -ldflags="-w -s" -o /app/service ./cmd/service

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/service /app/service

COPY config /app/config

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

USER appuser

EXPOSE 8080

CMD ["/app/service"]