# Stage 1: Build
FROM golang:1.24-alpine AS builder
WORKDIR /build

COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY svc/dissertation/ ./svc/dissertation/

RUN cd svc/dissertation && GOWORK=off CGO_ENABLED=0 GOOS=linux go build -o /bin/dissertation-server ./cmd/server

# Stage 2: Runtime
FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/dissertation-server /dissertation-server
EXPOSE 8086
ENTRYPOINT ["/dissertation-server"]
