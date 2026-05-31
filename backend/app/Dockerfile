# Stage 1: Build
FROM golang:1.24-alpine AS builder
WORKDIR /build

# Copy workspace + shared pkg + app module.
COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY app/ ./app/

RUN cd app && GOWORK=off CGO_ENABLED=0 GOOS=linux go build -o /bin/monolith-server ./cmd/server

# Stage 2: Runtime
FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/monolith-server /monolith-server
EXPOSE 8080
ENTRYPOINT ["/monolith-server"]
