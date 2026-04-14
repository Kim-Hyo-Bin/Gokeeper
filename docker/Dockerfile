# Build context: repository root.
#   docker build -f docker/Dockerfile -t gokeeper .
#
# Static binary (no CGO; glebarez/sqlite is pure Go)
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
	&& CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/gokeeper-keygen ./cmd/gokeeper-keygen

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata openssl \
	&& adduser -D -u 1000 -g 1000 appuser
WORKDIR /app
COPY --from=build /out/server /app/server
COPY --from=build /out/gokeeper-keygen /app/gokeeper-keygen
COPY docker/entrypoint.sh /app/entrypoint.sh
RUN chmod 755 /app/entrypoint.sh /app/gokeeper-keygen \
	&& mkdir -p /data \
	&& chown appuser:appuser /data /app/entrypoint.sh /app/gokeeper-keygen /app/server
USER appuser
ENV ADDR=:8080
ENV DATABASE_PATH=/data/licenses.db
# Key path / auto-generate defaults: docker/entrypoint.sh (exports LICENSE_PRIVATE_KEY_PATH before exec)
EXPOSE 8080
VOLUME ["/data"]
ENTRYPOINT ["/app/entrypoint.sh"]
