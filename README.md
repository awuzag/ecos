# ecos

`ecos`는 한국은행 경제통계시스템(ECOS) OpenAPI를 위한 독립 Go 클라이언트 라이브러리입니다.

SDK core를 작게 유지하면서 한국은행 ECOS 공식 OpenAPI 6개 서비스를 typed method로 제공합니다. OpenDART나 mwosa의 생성 산출물은 복사하지 않고, repo-local 문서에 기록한 ECOS API 조사 결과를 기준으로 구현합니다.

## 설치

```sh
go get github.com/awuzag/ecos
```

## 사용 예시

```go
package main

import (
	"context"
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

	result, err := client.Search(context.Background(), ecos.SearchRequest{
		Page:      ecos.NewPage(1, 10),
		StatCode:  ecos.StatCodeBankOfKoreaBaseRate,
		Cycle:     ecos.CycleMonthly,
		StartTime: "202001",
		EndTime:   "202604",
		ItemCodes: []ecos.ItemCode{ecos.ItemCodeBankOfKoreaBaseRate},
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, row := range result.Rows {
		log.Printf("%s %s", row.Time, row.Value)
	}
}
```

## 인증키

ECOS 인증키는 코드에 직접 쓰지 말고 환경변수로 주입합니다.

```sh
export ECOS_API_KEY="발급받은_인증키"
```

## 지원 범위

현재 저장소는 다음 범위를 포함합니다.

- root package `github.com/awuzag/ecos`
- `New(Config, ...Option)` 기반 client 생성
- `WithBaseURL`, `WithHTTPClient`, `WithTimeout` option
- ECOS path-style API 호출을 위한 내부 JSON request helper
- HTTP error, JSON decode error, ECOS business error 구분
- 공식 ECOS OpenAPI 6개 서비스 typed method
- `Cycle`, `Lang`, `Format`, `MessageCode`, `StatCode`, `ItemCode` typed const
- fake server 기반 기본 테스트와 `e2e` build tag 기반 live smoke 테스트

## API

| ECOS service | SDK method | 비고 |
| --- | --- | --- |
| `StatisticTableList` | `Tables` | 통계표 목록과 조회 가능 여부 |
| `StatisticItemList` | `Items` | 통계표별 항목, 주기, 제공 기간 |
| `StatisticSearch` | `Search` | 통계 시계열 조회 |
| `KeyStatisticList` | `KeyStatistics` | 주요 지표 최신값 |
| `StatisticMeta` | `Meta` | 통계 메타데이터 |
| `StatisticWord` | `Words` | 통계용어사전 |

`KeyStatistic.CYCLE` provider 필드는 SDK에서 `ReferenceTime`으로 노출합니다. 이 값은 주기 코드가 아니라 최신값 기준시점입니다.

## 개발

```sh
go mod tidy
go test ./...
go test -cover ./...
git diff --check
```

가능하면 아래 검증도 함께 수행합니다.

```sh
go test -race ./...
go vet ./...
```

기본 테스트는 live ECOS 호출을 하지 않습니다. 실제 ECOS 서버를 호출하는 e2e smoke는 별도 build tag와 `ECOS_API_KEY` 기반 optional workflow로 분리합니다.

```sh
go test -tags=e2e ./...
```

`Taskfile.yml`을 쓰면 로컬과 Docker 기반 검증을 같은 이름으로 실행할 수 있습니다.

```sh
task verify
task docker:verify
```

## 문서

- API 목록: `docs/apis/README.md`
- 공식 인벤토리: `docs/apis/official-inventory.md`
- OpenAPI 수집 계획: `docs/apis/openapi.md`
- typed SDK 체크리스트: `docs/apis/typed-sdk-checklist.md`
