# ecos

`ecos`는 한국은행 경제통계시스템(ECOS) OpenAPI를 위한 독립 Go 클라이언트 라이브러리입니다.

초기 목표는 SDK core를 작게 유지하면서 공식 ECOS API 목록, 요청 파라미터, 응답 필드를 새로 수집해 typed method로 확장하는 것입니다. OpenDART나 mwosa의 생성 산출물은 복사하지 않고, 문서화와 생성 흐름만 참고합니다.

## 설치

```sh
go get github.com/awuzag/ecos
```

## 사용 예시

```go
package main

import (
	"log"
	"os"

	"github.com/awuzag/ecos"
)

func main() {
	client, err := ecos.New(ecos.Config{
		APIKey: os.Getenv("ECOS_API_KEY"),
	})
	if err != nil {
		log.Fatal(err)
	}

	_ = client
}
```

## 인증키

ECOS 인증키는 코드에 직접 쓰지 말고 환경변수로 주입합니다.

```sh
export ECOS_API_KEY="발급받은_인증키"
```

## 지원 범위

현재 저장소는 다음 초기 뼈대를 포함합니다.

- root package `github.com/awuzag/ecos`
- `New(Config, ...Option)` 기반 client 생성
- `WithBaseURL`, `WithHTTPClient`, `WithTimeout` option
- ECOS path-style API 호출을 위한 내부 JSON request helper
- HTTP error, JSON decode error, ECOS business error 구분
- API 문서 수집과 typed SDK 구현을 위한 `docs/apis/` 템플릿

공식 API 목록과 typed method는 새로 수집한 문서를 기준으로 추가합니다.

## 개발

```sh
go mod tidy
go test ./...
git diff --check
```

가능하면 아래 검증도 함께 수행합니다.

```sh
go test -race ./...
go vet ./...
```

기본 테스트는 live ECOS 호출을 하지 않습니다. 실제 ECOS 서버를 호출하는 e2e smoke는 별도 build tag와 `ECOS_API_KEY` 기반 optional workflow로 분리합니다.

## 문서

- API 목록: `docs/apis/README.md`
- 공식 인벤토리: `docs/apis/official-inventory.md`
- OpenAPI 수집 계획: `docs/apis/openapi.md`
- typed SDK 체크리스트: `docs/apis/typed-sdk-checklist.md`
