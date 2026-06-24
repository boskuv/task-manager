# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/task-manager \
    ./cmd/api

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata \
    && adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /out/task-manager /app/task-manager
COPY configs/config.example.yaml /app/configs/config.yaml

USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/task-manager"]
