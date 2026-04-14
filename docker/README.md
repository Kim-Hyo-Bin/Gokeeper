# GoKeeper — Docker

Build from the repository root:

```bash
docker build -f docker/Dockerfile -t gokeeper .
```

## Run (default)

SQLite database and Ed25519 key pair under `/data` on first start:

```bash
docker run -p 8080:8080 -v gokeeper-data:/data gokeeper
```

- Private key: `/data/license_private.pem`
- Public key (for client apps): `/data/license_public.pem`
- Use a named or bind-mounted volume so the same keys and DB survive restarts.

## Disable auto key generation

Use when you always mount an existing key file:

```bash
docker run -p 8080:8080 \
  -v "$(pwd)/keys:/keys" \
  -e LICENSE_AUTO_GENERATE_KEY=false \
  -e LICENSE_PRIVATE_KEY_PATH=/keys/license_private.pem \
  gokeeper
```

## Inline PEM

The entrypoint skips file generation and path-based load when `LICENSE_PRIVATE_KEY_PEM` is set (it unsets `LICENSE_PRIVATE_KEY_PATH` before starting the server):

```bash
docker run ... -e LICENSE_PRIVATE_KEY_PEM="$(cat license_private.pem)" gokeeper
```

## Key pair CLI (`gokeeper-keygen`)

The image ships `/app/gokeeper-keygen` (same as `go run ./cmd/gokeeper-keygen` locally). Override the entrypoint to generate keys inside a volume, for example:

```bash
docker run --rm --entrypoint /app/gokeeper-keygen \
  -v gokeeper-data:/data \
  gokeeper \
  -private-out /data/license_private.pem -public-out /data/license_public.pem
```

Then run the server as usual with that volume mounted.

## Build context

The **context must be the repository root** (the `.` at the end). Paths like `go.mod` and `docker/entrypoint.sh` are resolved from that directory.

```bash
# Good (from repo root)
docker build -f docker/Dockerfile -t gokeeper .

# Wrong — context is only docker/, so go.mod and paths are missing
cd docker && docker build -t gokeeper .
```

`.dockerignore` at the repo root applies to that context.
