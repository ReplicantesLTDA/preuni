# Stage 1: Build
FROM golang:1.23-alpine AS builder
WORKDIR /build

COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY svc/notification/ ./svc/notification/

RUN cd svc/notification && CGO_ENABLED=0 GOOS=linux go build -o /bin/notification-server ./cmd/server

# Stage 2: Runtime
FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/notification-server /notification-server
EXPOSE 8087
ENTRYPOINT ["/notification-server"]
