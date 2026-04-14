# Build context must be the repository root:
#   docker build -f docker/Dockerfile -t gokeeper .
#
# Static binaries (no CGO; glebarez/sqlite is pure Go).
FROM golang:1.23-alpine AS build
RUN apk add --no-cache git ca-certificates
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# .git is usually dockerignored; disable VCS stamping so the build does not fail.
ENV GOFLAGS=-buildvcs=false
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
	&& CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/gokeeper-keygen ./cmd/gokeeper-keygen

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata openssl \
	&& adduser -D -u 1000 -g 1000 appuser
WORKDIR /app
COPY --from=build /out/server /app/server
COPY --from=build /out/gokeeper-keygen /app/gokeeper-keygen
# From build image (after `COPY . .`), not host path — avoids failures when context is wrong.
COPY --from=build /src/docker/entrypoint.sh /app/entrypoint.sh
RUN chmod 755 /app/entrypoint.sh /app/gokeeper-keygen /app/server \
	&& mkdir -p /data \
	&& chown appuser:appuser /data /app/entrypoint.sh /app/gokeeper-keygen /app/server
USER appuser
ENV ADDR=:8080
ENV DATABASE_PATH=/data/licenses.db
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/app/entrypoint.sh"]
