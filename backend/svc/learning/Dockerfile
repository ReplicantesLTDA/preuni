# Stage 1: Build
FROM golang:1.23-alpine AS builder
WORKDIR /build

COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY svc/learning/ ./svc/learning/

RUN cd svc/learning && CGO_ENABLED=0 GOOS=linux go build -o /bin/learning-server ./cmd/server

# Stage 2: Runtime
FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/learning-server /learning-server
EXPOSE 8084
ENTRYPOINT ["/learning-server"]
