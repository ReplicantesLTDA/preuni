# Stage 1: Build
FROM golang:1.23-alpine AS builder
WORKDIR /build

COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY svc/simulation/ ./svc/simulation/

RUN cd svc/simulation && CGO_ENABLED=0 GOOS=linux go build -o /bin/simulation-server ./cmd/server

# Stage 2: Runtime
FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/simulation-server /simulation-server
EXPOSE 8085
ENTRYPOINT ["/simulation-server"]
