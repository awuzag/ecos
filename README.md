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

CLI는 인증키를 `--api-key`, `ECOS_API_KEY`, repo-local `.env` 순서로 읽습니다. 인증키 값은 출력하지 않습니다.

repo-local `.env`를 사용할 때는 다음처럼 환경변수 이름만 둡니다.

```sh
ECOS_API_KEY=발급받은_인증키
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
- Cobra 기반 `cmd/ecos` CLI
- fake server 기반 기본 테스트와 `e2e` build tag 기반 live smoke 테스트

## CLI

개발 체크아웃에서는 `go run`으로 바로 실행할 수 있습니다.

```sh
go run ./cmd/ecos show key-statistics --start 1 --end 5
go run ./cmd/ecos show key-statistics --start 1 --end 5 --json
```

설치해서 쓰려면 태그 기준으로 `cmd/ecos`를 설치합니다.

```sh
go install github.com/awuzag/ecos/cmd/ecos@latest
```

명령 이름은 verb first 형태를 사용합니다.

```sh
ecos list tables --start 1 --end 20
ecos list items --stat-code 722Y001 --start 1 --end 20
ecos search observations --stat-code 722Y001 --cycle M --time-start 202001 --time-end 202604 --item-code 0101000
ecos show key-statistics --start 1 --end 5 --json
ecos show word 소비자동향지수
ecos show meta 경제심리지수
```

공통 옵션은 `--api-key`, `--env-file`, `--base-url`, `--start`, `--end`, `--json`입니다. `--start`와 `--end`는 ECOS path-style row 범위이고, `search observations`의 기간은 `--time-start`, `--time-end`로 지정합니다.

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
