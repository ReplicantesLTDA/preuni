# Stage 1: Build
FROM golang:1.24-alpine AS builder
WORKDIR /build

# Copy workspace + shared packages + dependency modules.
# Monolith replace-directives auth/user/pkg via relative paths, so all
# four directories must be present in the build context.
COPY go.work go.work.sum ./
COPY pkg/ ./pkg/
COPY svc/auth/ ./svc/auth/
COPY svc/user/ ./svc/user/
COPY svc/monolith/ ./svc/monolith/

RUN cd svc/monolith && GOWORK=off CGO_ENABLED=0 GOOS=linux go build -o /bin/monolith-server ./cmd/server

# Stage 2: Runtime
FROM gcr.io/distroless/static-debian12
COPY --from=builder /bin/monolith-server /monolith-server
EXPOSE 8080
ENTRYPOINT ["/monolith-server"]
