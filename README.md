# GoKeeper

License issuance and **offline validation** using **Ed25519** asymmetric keys. The server holds the private key and signs license payloads; clients verify signatures with the embedded or distributed **public key** (PEM).

Stack: **Gin**, **GORM**, embedded **SQLite** (pure Go driver, no CGO), Go **1.23+**.

## Features

- Issue licenses with a **UUID** identity; optional expiry (`exp` Unix timestamp; `0` = no expiry).
- Persist license metadata in **SQLite** (path configurable).
- **`internal/service`**: issue, fetch, revoke, and **online verification** (signature + DB revocation/expiry) for use from HTTP or tests.
- **Client package** [`gokeeper/pkg/license`](pkg/license): `Validate` (offline, includes expiry) and `VerifySignature` (signature only, for advanced flows).
- **`gokeeper-keygen`**: CLI to generate PKCS#8 private + public PEM files ([`cmd/gokeeper-keygen`](cmd/gokeeper-keygen)).
- **Docker** image includes the server and `gokeeper-keygen`; entrypoint can generate a key pair on first start under `/data` (see [docker/README.md](docker/README.md)).
- **GitHub Actions**: `gofmt`, `go test -race`, **golangci-lint**.

## Quick start (local)

```bash
go run ./cmd/server
```

By default the server expects a private key. For local development you can auto-generate keys next to the database directory:

```bash
export LICENSE_GENERATE_KEYS_DEV=true
go run ./cmd/server
```

Or generate keys explicitly (same format as the server expects):

```bash
go run ./cmd/gokeeper-keygen -private-out ./data/license_private.pem -public-out ./data/license_public.pem
export LICENSE_PRIVATE_KEY_PATH=./data/license_private.pem
go run ./cmd/server
```

## Configuration (environment)

| Variable | Description |
|----------|-------------|
| `ADDR` | Listen address (default `:8080`). |
| `DATABASE_PATH` | SQLite file path (default `./data/licenses.db`). |
| `LICENSE_PRIVATE_KEY_PATH` | Path to Ed25519 **private** key (PKCS#8 PEM). |
| `LICENSE_PRIVATE_KEY_PEM` | Inline private key PEM (e.g. secret injection); path is ignored if this is set. |
| `LICENSE_GENERATE_KEYS_DEV` | If `true`, generate `license_private.pem` / `license_public.pem` beside the DB dir when missing (dev-oriented). |
| `AUTO_MIGRATE` | If `true`, run GORM auto-migrate on start (default `true`). |

Docker-specific variables (`LICENSE_AUTO_GENERATE_KEY`, entrypoint behavior) are documented in [docker/README.md](docker/README.md).

## HTTP API

Base URL: `http://<host>:8080`

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Liveness. |
| `POST` | `/v1/licenses` | Issue a license. Optional JSON body: `{"expires_at": "2026-12-31T00:00:00Z"}` (RFC3339). Empty body = no expiry. Response includes `id` (UUID) and `license_key`. |
| `POST` | `/v1/licenses/verify` | Online check: JSON `{"license_key":"..."}`. Verifies signature with the server key, then DB row, revocation, and expiry. Response: `valid`, `reason` (`ok`, `revoked`, `expired`, `unknown`, `mismatch`, `invalid_signature`, …), and optional metadata. |
| `GET` | `/v1/licenses/:id` | Metadata and stored `license_key`. |
| `POST` | `/v1/licenses/:id/revoke` | Sets `revoked_at` in the database. |

There is **no** HTTP API to change an existing license in place. To change expiry or rotate a key, **revoke** the old id and **issue** a new license.

### curl examples

Set `BASE` to your server (e.g. `http://localhost:8080`).

**Issue a 7-day license** (GNU `date`, Linux / WSL):

```bash
BASE=http://localhost:8080
EXP=$(date -u -d '+7 days' '+%Y-%m-%dT%H:%M:%SZ')
curl -sS -X POST "$BASE/v1/licenses" \
  -H 'Content-Type: application/json' \
  -d "{\"expires_at\":\"$EXP\"}"
```

On macOS (BSD `date`):

```bash
EXP=$(date -u -v+7d '+%Y-%m-%dT%H:%M:%SZ')
```

**Issue with no expiry** (empty body):

```bash
curl -sS -X POST "$BASE/v1/licenses"
```

**Get** (replace `LICENSE_ID` with the `id` from the issue response):

```bash
curl -sS "$BASE/v1/licenses/LICENSE_ID"
```

**Verify online** (replace `LICENSE_KEY`):

```bash
curl -sS -X POST "$BASE/v1/licenses/verify" \
  -H 'Content-Type: application/json' \
  -d '{"license_key":"LICENSE_KEY"}'
```

**Revoke**:

```bash
curl -sS -X POST "$BASE/v1/licenses/LICENSE_ID/revoke"
```

**Offline vs online:** Client `Validate` checks **signature and expiry** only; it does not see revocation. **When the app can reach the license server, prefer online checks** via **`POST /v1/licenses/verify`** (or `internal/service.License.Verify`) so revocation and DB state apply. Offline validation exists mainly for **air-gapped (closed-network)** deployments that cannot call the server, using a distributed public key. See [examples/](examples/) for snippets in several languages.

## Client library

```go
import "gokeeper/pkg/license"

claims, err := license.Validate(licenseKeyString, publicKeyPEMBytes)
if err != nil {
    // handle license.ErrExpired, license.ErrVerify, etc.
}
// claims.LicenseID, claims.Exp
```

The license string format is `base64url(payload).base64url(signature)` where the payload is JSON `{"license_id":"...","exp":...}`.

Multi-language **API and offline validation snippets** live under [examples/](examples/).

## Docker

From the repository root:

```bash
docker build -f docker/Dockerfile -t gokeeper .
docker run -p 8080:8080 -v gokeeper-data:/data gokeeper
```

See [docker/README.md](docker/README.md) for key auto-generation, `gokeeper-keygen` in the image, disabling auto-generation, and inline PEM.

## Development

```bash
go test ./... -race
golangci-lint run ./...
gofmt -w .
```

CI runs on pushes and pull requests to `main` / `master` (see [.github/workflows/ci.yml](.github/workflows/ci.yml)).

## Module path

The Go module is `gokeeper`. If you publish this repository, rename the module in `go.mod` to your canonical path (for example `github.com/<org>/gokeeper`) and update imports accordingly.

## License format & security

- Keys are **Ed25519**; private key must be stored securely (filesystem permissions, secrets manager, or mounted volume).
- Distribute only the **public** PEM to client applications.
- Treat issued `license_key` values as bearer credentials when stored or logged.

---

Korean documentation: [README_kr.md](README_kr.md).
