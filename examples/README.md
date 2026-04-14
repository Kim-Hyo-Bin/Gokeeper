# GoKeeper client examples

Small, copy-paste-friendly samples for calling the HTTP API and (where noted) **offline** Ed25519 license validation. Wire format matches [`pkg/license`](../pkg/license): `base64url(JSON).base64url(signature)`.

**Why both online and offline samples?** Some deployments run on **air-gapped (closed) networks** where clients cannot call the license server; there you validate only with the **embedded public key** and the signed `license_key` (offline). **Whenever clients can reach the license server, prefer online verification** (`POST /v1/licenses/verify`): it reflects **revocation** and stays aligned with server-side state. Offline remains a fallback for closed networks. The examples cover both so you can pick the mode that fits each site.

## Layout

| Directory | Contents |
|-----------|----------|
| [go/](go/) | Go: REST client + offline validate via `gokeeper/pkg/license` |
| [python/](python/) | `requests` API client + `cryptography` offline validate |
| [javascript/](javascript/) | Node.js `fetch` API client + `crypto` offline validate |
| [java/](java/) | Java 17+ `HttpClient` API client + offline validate |

## Common environment

- **`BASE_URL`** — server root (default `http://localhost:8080`).
- **`PUBLIC_KEY_PATH`** — path to **Ed25519 public key PEM** (offline samples).
- **`LICENSE_KEY`** — issued `license_key` string (offline samples).

## Run hints

```bash
# Go (module root examples/go; replace points at repo root)
cd examples/go && go mod tidy
cd examples/go && LICENSE_KEY='...' PUBLIC_KEY_PATH='...' go run ./offline
cd examples/go && go run ./api

# Python
cd examples/python && pip install -r requirements.txt
python api_client.py
python offline_validate.py

# JavaScript (Node 18+)
cd examples/javascript && node api_client.mjs
node offline_validate.mjs

# Java
cd examples/java && javac *.java && java ApiExample
cd examples/java && java OfflineValidate
```

Adjust paths and environment variables for your deployment.
