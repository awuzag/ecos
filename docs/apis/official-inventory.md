# 공식 API 인벤토리

- 확인 날짜: 미수집
- 공식 문서 URL: https://ecos.bok.or.kr/api/
- 공식 개발가이드 API 수: 미수집

## 공통 요청/응답 규칙

- 모든 API는 인증키를 요구한다.
- 기본 인증키 환경변수는 `ECOS_API_KEY`다.
- 기본 요청 형식은 `json`, 기본 언어는 `kr`로 둔다.
- 인증키는 URL path segment에 포함되므로 error, log, fixture에 남기지 않는다.
- 기본 테스트는 live 호출 없이 fake server로 검증한다.
- business error는 ECOS provider-native code/message를 보존한다.

## 파라미터 프로필

| Profile | 요청 파라미터 | 비고 |
| --- | --- | --- |
| `paged` | `start`, `end` | ECOS path의 요청 시작/종료 건수 |
| `statistic-table` | `stat_code` | 통계표 목록 또는 항목 조회 |
| `statistic-search` | `stat_code`, `cycle`, `start_time`, `end_time`, `item_code...` | 통계 시계열 조회 |

## API 인벤토리 초안

| Service | 용도 | 필수 segment | 응답 root | SDK method |
| --- | --- | --- | --- | --- |
| `StatisticTableList` | 통계표 목록 | 미수집 | 미수집 | 미정 |
| `StatisticWord` | 통계용어사전 | 미수집 | 미수집 | 미정 |
| `StatisticItemList` | 통계항목 목록 | 미수집 | 미수집 | 미정 |
| `StatisticSearch` | 통계 조회 | 미수집 | 미수집 | 미정 |
| `KeyStatisticList` | 주요 통계지표 | 미수집 | 미수집 | 미정 |
| `StatisticMeta` | 통계 메타데이터 | 미수집 | 미수집 | 미정 |

## 추적 기준

- `Service`, `Params`, `Response fields`는 공식 문서에서 확인한 내용만 확정값으로 적는다.
- 블로그나 예제 코드는 공식 문서가 비어 있을 때만 보조 근거로 사용한다.
- SDK typed method 대응표는 `docs/apis/typed-sdk-checklist.md`에서 별도로 추적한다.
- 전체 요청/응답 필드 스키마는 추후 OpenAPI 산출물에서 확인한다.
