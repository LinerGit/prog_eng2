FROM golang:1.25.7-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/ticket-service ./cmd

FROM alpine:3.22

RUN adduser -D -H appuser
WORKDIR /app
COPY --from=builder /out/ticket-service /app/ticket-service
COPY migrations /app/migrations

USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/ticket-service"]
