# SDK typed method 체크리스트

- 기준 문서: `docs/apis/official-inventory.md`
- 확인 날짜: 2026-05-31
- 범위: 한국은행 ECOS 공식 OpenAPI 전체 service
- SDK package: `github.com/awuzag/ecos`

## 요약

| Group | APIs | Typed methods |
| --- | ---: | ---: |
| statistics | 2 | 2 |
| metadata | 4 | 4 |
| total | 6 | 6 |

## 대응표

| Group | Service | Endpoint shape | SDK method | 상태 |
| --- | --- | --- | --- | --- |
| statistics | `StatisticSearch` | `/api/StatisticSearch/{apiKey}/json/kr/{start}/{end}/{stat_code}/{cycle}/{start_time}/{end_time}/{item_code1}/{item_code2}/{item_code3}/{item_code4}` | `Search(ctx, SearchRequest)` | 구현 |
| statistics | `KeyStatisticList` | `/api/KeyStatisticList/{apiKey}/json/kr/{start}/{end}` | `KeyStatistics(ctx, KeyStatisticsRequest)` | 구현 |
| metadata | `StatisticTableList` | `/api/StatisticTableList/{apiKey}/json/kr/{start}/{end}/{stat_code?}` | `Tables(ctx, TablesRequest)` | 구현 |
| metadata | `StatisticItemList` | `/api/StatisticItemList/{apiKey}/json/kr/{start}/{end}/{stat_code}` | `Items(ctx, ItemsRequest)` | 구현 |
| metadata | `StatisticWord` | `/api/StatisticWord/{apiKey}/json/kr/{start}/{end}/{word}` | `Words(ctx, WordsRequest)` | 구현 |
| metadata | `StatisticMeta` | `/api/StatisticMeta/{apiKey}/json/kr/{start}/{end}/{data_name}` | `Meta(ctx, MetaRequest)` | 구현 |

## 이름 결정 기준

- 공식 service 이름은 source traceability를 위해 문서와 생성 source에 보존한다.
- public SDK method는 Go 사용자에게 자연스러운 이름으로 둔다.
- friendly 이름이 기계적으로 도출되지 않으면 별도 override 문서에 둔다.
- 기존 public method 이름을 바꾸면 changelog와 migration note를 남긴다.

## 구현 메모

- 응답 row의 provider-native 필드명은 JSON tag에 보존한다.
- `KeyStatisticList.CYCLE`은 SDK에서 `ReferenceTime` 필드로 노출한다.
- `StatisticSearch`에서 비어 있는 item code segment는 `?`를 URL path segment로 escaped 처리한다.
- `StatisticMeta`, `StatisticWord`의 한글 path segment는 `url.PathEscape` 결과로 전송한다.
- 기본 테스트는 fake HTTP server 기반이며 live 호출은 `e2e` build tag와 `ECOS_API_KEY` 조건에서만 실행한다.
