# GoKeeper

**Ed25519** 비대칭키로 라이선스를 발급하고 **오프라인에서 검증**하는 서버·클라이언트 구성입니다. 서버는 **개인키**로 페이로드를 서명하고, 클라이언트는 배포된 **공개키(PEM)**로 서명만 확인합니다.

스택: **Gin**, **GORM**, 임베디드 **SQLite**(순수 Go 드라이버, CGO 없음), Go **1.23+**.

## 주요 기능

- 라이선스 ID는 **UUID**; 선택적 만료(`exp` Unix 시각, `0`이면 무기한).
- 라이선스 메타데이터를 **SQLite**에 저장(경로 설정 가능).
- **`internal/service`**: 발급·조회·폐기·**온라인 검증**(서명 + DB 폐기/만료)을 HTTP·테스트에서 재사용.
- **클라이언트 패키지** [`gokeeper/pkg/license`](pkg/license): 오프라인 `Validate`(만료 포함), 서명만 확인하는 `VerifySignature`.
- **`gokeeper-keygen`**: PKCS#8 개인키·공개키 PEM 생성 CLI ([`cmd/gokeeper-keygen`](cmd/gokeeper-keygen)).
- **Docker** 이미지에 서버와 `gokeeper-keygen` 포함; 첫 기동 시 `/data`에 키쌍 자동 생성 가능([docker/README.md](docker/README.md)).
- **GitHub Actions**: `gofmt`, `go test -race`, **golangci-lint**.

## 로컬 실행

```bash
go run ./cmd/server
```

기본적으로 개인키가 필요합니다. 개발용으로 DB 디렉터리 옆에 키를 자동 생성하려면:

```bash
export LICENSE_GENERATE_KEYS_DEV=true
go run ./cmd/server
```

또는 키를 명시적으로 생성한 뒤(서버와 동일 형식):

```bash
go run ./cmd/gokeeper-keygen -private-out ./data/license_private.pem -public-out ./data/license_public.pem
export LICENSE_PRIVATE_KEY_PATH=./data/license_private.pem
go run ./cmd/server
```

## 환경 변수

| 변수 | 설명 |
|------|------|
| `ADDR` | 리스닝 주소 (기본 `:8080`). |
| `DATABASE_PATH` | SQLite 파일 경로 (기본 `./data/licenses.db`). |
| `LICENSE_PRIVATE_KEY_PATH` | Ed25519 **개인키** 파일 경로 (PKCS#8 PEM). |
| `LICENSE_PRIVATE_KEY_PEM` | 인라인 개인키 PEM(시크릿 주입 등). 설정 시 경로보다 우선. |
| `LICENSE_GENERATE_KEYS_DEV` | `true`이면 키가 없을 때 DB 디렉터리 옆에 `license_private.pem` / `license_public.pem` 생성(개발용). |
| `AUTO_MIGRATE` | `true`이면 기동 시 GORM AutoMigrate (기본 `true`). |

Docker 전용 변수는 [docker/README.md](docker/README.md)를 참고하세요.

## HTTP API

기본 URL: `http://<호스트>:8080`

| 메서드 | 경로 | 설명 |
|--------|------|------|
| `GET` | `/health` | 헬스 체크. |
| `POST` | `/v1/licenses` | 라이선스 발급. 선택 JSON: `{"expires_at": "2026-12-31T00:00:00Z"}` (RFC3339). 본문 없음 = 만료 없음. 응답에 `id`(UUID), `license_key`. |
| `POST` | `/v1/licenses/verify` | 온라인 검증. JSON `{"license_key":"..."}`. 서버 키로 서명 확인 후 DB 행·폐기·만료 반영. 응답: `valid`, `reason`(`ok`, `revoked`, `expired`, `unknown`, `mismatch`, `invalid_signature` 등), 선택 필드. |
| `GET` | `/v1/licenses/:id` | 메타데이터 및 저장된 `license_key`. |
| `POST` | `/v1/licenses/:id/revoke` | DB에 `revoked_at` 설정. |

이미 발급된 라이선스를 **HTTP로 수정**하는 API는 없습니다. 만료일을 바꾸거나 키를 갈아끼우려면 기존 건을 **폐기(revoke)** 한 뒤 **새로 발급(issue)** 하세요.

### curl 예시

`BASE`를 서버 주소로 맞춥니다 (예: `http://localhost:8080`).

**7일짜리 라이선스 발급** (GNU `date`, Linux / WSL):

```bash
BASE=http://localhost:8080
EXP=$(date -u -d '+7 days' '+%Y-%m-%dT%H:%M:%SZ')
curl -sS -X POST "$BASE/v1/licenses" \
  -H 'Content-Type: application/json' \
  -d "{\"expires_at\":\"$EXP\"}"
```

macOS (BSD `date`):

```bash
EXP=$(date -u -v+7d '+%Y-%m-%dT%H:%M:%SZ')
```

**만료 없이 발급** (빈 본문):

```bash
curl -sS -X POST "$BASE/v1/licenses"
```

**조회** (`LICENSE_ID`는 발급 응답의 `id`):

```bash
curl -sS "$BASE/v1/licenses/LICENSE_ID"
```

**온라인 검증** (`LICENSE_KEY` 교체):

```bash
curl -sS -X POST "$BASE/v1/licenses/verify" \
  -H 'Content-Type: application/json' \
  -d '{"license_key":"LICENSE_KEY"}'
```

**폐기(제거)**:

```bash
curl -sS -X POST "$BASE/v1/licenses/LICENSE_ID/revoke"
```

**오프라인 vs 온라인:** 클라이언트 `Validate`는 **서명·만료**만 검사하며 폐기 여부는 모릅니다. **라이선스 서버에 연결할 수 있으면 온라인 검증을 권장합니다** — **`POST /v1/licenses/verify`** 또는 `internal/service.License.Verify`로 **폐기**와 DB 상태를 반영할 수 있습니다. 오프라인 검증은 주로 **폐쇄망**처럼 서버에 닿을 수 없을 때를 위한 **대안**입니다(배포한 공개키 + `license_key`). 두 방식을 모두 지원하는 이유는 폐쇄망 배포를 고려했기 때문입니다. 여러 언어 예시는 [examples/](examples/) 를 참고하세요.

## 클라이언트 라이브러리

```go
import "gokeeper/pkg/license"

claims, err := license.Validate(licenseKeyString, publicKeyPEMBytes)
if err != nil {
    // license.ErrExpired, license.ErrVerify 등 처리
}
// claims.LicenseID, claims.Exp
```

라이선스 문자열 형식은 `base64url(페이로드).base64url(서명)` 이며, 페이로드는 JSON `{"license_id":"...","exp":...}` 입니다.

여러 언어의 **API·오프라인 검증 예시 코드**는 [examples/](examples/) 를 참고하세요.

## Docker

저장소 루트에서:

```bash
docker build -f docker/Dockerfile -t gokeeper .
docker run -p 8080:8080 -v gokeeper-data:/data gokeeper
```

키 자동 생성, 이미지 내 `gokeeper-keygen`, 인라인 PEM 등은 [docker/README.md](docker/README.md)를 참고하세요.

## 개발

```bash
go test ./... -race
golangci-lint run ./...
gofmt -w .
```

CI는 `main` / `master`에 대한 push·PR에서 실행됩니다([.github/workflows/ci.yml](.github/workflows/ci.yml)).

## 모듈 경로

Go 모듈명은 `gokeeper`입니다. 원격 저장소에 올릴 때는 `go.mod`을 예를 들어 `github.com/<조직>/gokeeper` 형태로 바꾸고 import 경로를 맞추면 됩니다.

## 보안

- 키는 **Ed25519**; 개인키는 파일 권한·시크릿 매니저·볼륨 마운트 등으로 보호하세요.
- 클라이언트에는 **공개키 PEM**만 배포하세요.
- 발급된 `license_key`는 로그·저장 시 민감 정보로 취급하는 것이 좋습니다.

---

영문 문서: [README.md](README.md).
