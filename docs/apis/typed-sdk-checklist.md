# SDK typed method 체크리스트

- 기준 문서: `docs/apis/official-inventory.md`
- 확인 날짜: 미수집
- 범위: 한국은행 ECOS 공식 OpenAPI 전체 service
- SDK package: `github.com/awuzag/ecos`

## 요약

| Group | APIs | Typed methods |
| --- | ---: | ---: |
| statistics | 미수집 | 0 |
| metadata | 미수집 | 0 |
| total | 미수집 | 0 |

## 대응표

| Group | Service | Endpoint shape | SDK method | 상태 |
| --- | --- | --- | --- | --- |
| statistics | `StatisticSearch` | `/api/StatisticSearch/{apiKey}/json/kr/...` | 미정 | 미구현 |
| statistics | `KeyStatisticList` | `/api/KeyStatisticList/{apiKey}/json/kr/...` | 미정 | 미구현 |
| metadata | `StatisticTableList` | `/api/StatisticTableList/{apiKey}/json/kr/...` | 미정 | 미구현 |
| metadata | `StatisticItemList` | `/api/StatisticItemList/{apiKey}/json/kr/...` | 미정 | 미구현 |
| metadata | `StatisticWord` | `/api/StatisticWord/{apiKey}/json/kr/...` | 미정 | 미구현 |
| metadata | `StatisticMeta` | `/api/StatisticMeta/{apiKey}/json/kr/...` | 미정 | 미구현 |

## 이름 결정 기준

- 공식 service 이름은 source traceability를 위해 문서와 생성 source에 보존한다.
- public SDK method는 Go 사용자에게 자연스러운 이름으로 둔다.
- friendly 이름이 기계적으로 도출되지 않으면 별도 override 문서에 둔다.
- 기존 public method 이름을 바꾸면 changelog와 migration note를 남긴다.
