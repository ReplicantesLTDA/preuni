# Stage 1: Build
FROM golang:1.24-alpine AS builder
# go.work/go.mod pin toolchain go1.26.6 (security patches) ahead of this
# base image's Go 1.24 -- official Go images set GOTOOLCHAIN=local, which
# blocks the usual auto-fetch and breaks the build on every toolchain bump.
# auto restores that auto-fetch so this Dockerfile never needs editing
# again just because go.work's toolchain line changes.
ENV GOTOOLCHAIN=auto
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
