# Go examples

`go.mod` uses `replace gokeeper => ../..` so `gokeeper/pkg/license` resolves to this repository without publishing the module.

```bash
go mod tidy
LICENSE_KEY='...' PUBLIC_KEY_PATH='path/to/license_public.pem' go run ./offline
BASE_URL=http://localhost:8080 go run ./api
```
