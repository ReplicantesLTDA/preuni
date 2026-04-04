# Stage 1: Build
FROM golang:1.23-alpine AS builder
WORKDIR /build

COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY svc/content/ ./svc/content/

RUN cd svc/content && CGO_ENABLED=0 GOOS=linux go build -o /bin/content-server ./cmd/server

# Stage 2: Runtime
FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/content-server /content-server
EXPOSE 8083
ENTRYPOINT ["/content-server"]
