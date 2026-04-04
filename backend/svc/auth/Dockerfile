# Stage 1: Build
FROM golang:1.24-alpine AS builder
WORKDIR /build

# Copy workspace definition and shared packages
COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY svc/auth/ ./svc/auth/

RUN cd svc/auth && GOWORK=off CGO_ENABLED=0 GOOS=linux go build -o /bin/auth-server ./cmd/server

# Stage 2: Runtime — minimal distroless image
FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/auth-server /auth-server
EXPOSE 8081
ENTRYPOINT ["/auth-server"]
