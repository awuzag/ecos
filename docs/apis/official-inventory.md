# 공식 API 인벤토리

- 확인 날짜: 2026-05-31
- 공식 문서 URL: https://ecos.bok.or.kr/api/
- 공식 개발가이드 API 수: 6
- 조사 노트: `docs/apis/ecos-openapi-research.md`

## 공통 요청/응답 규칙

- 모든 API는 인증키를 요구한다.
- 기본 인증키 환경변수는 `ECOS_API_KEY`다.
- 기본 요청 형식은 `json`, 기본 언어는 `kr`로 둔다.
- 인증키는 URL path segment에 포함되므로 error, log, fixture에 남기지 않는다.
- 공통 호출 형식은 `https://ecos.bok.or.kr/api/{service}/{apiKey}/{format}/{lang}/{start}/{end}/...` 이다.
- 기본 테스트는 live 호출 없이 fake server로 검증한다.
- business error는 ECOS provider-native code/message를 보존한다.
- JSON 응답은 대체로 서비스명을 root key로 두고 내부에 `list_total_count`, `row`를 둔다.
- `KeyStatisticList`는 `row_count`도 함께 내려온다.

## 파라미터 프로필

| Profile | 요청 파라미터 | 비고 |
| --- | --- | --- |
| `paged` | `start`, `end` | ECOS path의 요청 시작/종료 건수 |
| `statistic-table` | `stat_code` | 통계표 목록 또는 항목 조회 |
| `statistic-search` | `stat_code`, `cycle`, `start_time`, `end_time`, `item_code...` | 통계 시계열 조회 |
| `word` | `word` | 통계용어사전 검색어. URL escaping 필요 |
| `metadata` | `data_name` | 통계메타DB 데이터명. URL escaping 필요 |

## API 인벤토리

| Service | 용도 | 필수 segment | 응답 root | SDK method |
| --- | --- | --- | --- | --- |
| `StatisticTableList` | 통계표 목록 | `{service}/{apiKey}/{format}/{lang}/{start}/{end}/{stat_code?}` | `StatisticTableList` | 미정 |
| `StatisticItemList` | 통계항목 목록 | `{service}/{apiKey}/{format}/{lang}/{start}/{end}/{stat_code}` | `StatisticItemList` | 미정 |
| `StatisticSearch` | 통계 조회 | `{service}/{apiKey}/{format}/{lang}/{start}/{end}/{stat_code}/{cycle}/{start_time}/{end_time}/{item_code1}/{item_code2}/{item_code3}/{item_code4}` | `StatisticSearch` | 미정 |
| `KeyStatisticList` | 주요 통계지표 | `{service}/{apiKey}/{format}/{lang}/{start}/{end}` | `KeyStatisticList` | 미정 |
| `StatisticMeta` | 통계 메타데이터 | `{service}/{apiKey}/{format}/{lang}/{start}/{end}/{data_name}` | `StatisticMeta` | 미정 |
| `StatisticWord` | 통계용어사전 | `{service}/{apiKey}/{format}/{lang}/{start}/{end}/{word}` | `StatisticWord` | 미정 |

## 서비스별 응답 필드

### `StatisticTableList`

| Field | 의미 |
| --- | --- |
| `P_STAT_CODE` | 상위 통계 코드 |
| `STAT_CODE` | 통계표 코드 |
| `STAT_NAME` | 통계표 이름 |
| `CYCLE` | 대표 주기. 계층 노드이면 `null` 가능 |
| `SRCH_YN` | 실제 조회 가능 여부 |
| `ORG_NAME` | 작성 기관 |

### `StatisticItemList`

| Field | 의미 |
| --- | --- |
| `STAT_CODE` | 통계표 코드 |
| `STAT_NAME` | 통계표 이름 |
| `GRP_CODE` | 항목 그룹 코드 |
| `GRP_NAME` | 항목 그룹 이름 |
| `ITEM_CODE` | 항목 코드 |
| `ITEM_NAME` | 항목 이름 |
| `P_ITEM_CODE` | 상위 항목 코드 |
| `P_ITEM_NAME` | 상위 항목 이름 |
| `CYCLE` | 항목 제공 주기 |
| `START_TIME` | 제공 시작 시점 |
| `END_TIME` | 제공 종료 시점 |
| `DATA_CNT` | 데이터 개수 |
| `UNIT_NAME` | 단위 |
| `WEIGHT` | 가중치 |

### `StatisticSearch`

| Field | 의미 |
| --- | --- |
| `STAT_CODE` | 통계표 코드 |
| `STAT_NAME` | 통계표 이름 |
| `ITEM_CODE1` ~ `ITEM_CODE4` | 항목 코드 |
| `ITEM_NAME1` ~ `ITEM_NAME4` | 항목 이름 |
| `UNIT_NAME` | 단위 |
| `WGT` | 가중치 |
| `TIME` | 관측 시점 |
| `DATA_VALUE` | 값 |

### `KeyStatisticList`

| Field | 의미 |
| --- | --- |
| `CLASS_NAME` | 지표 분류 |
| `KEYSTAT_NAME` | 지표명 |
| `DATA_VALUE` | 최신값 |
| `CYCLE` | 최신값 기준시점 |
| `UNIT_NAME` | 단위 |

### `StatisticMeta`

| Field | 의미 |
| --- | --- |
| `LVL` | 메타 항목 계층 레벨 |
| `P_CONT_CODE` | 상위 콘텐츠 코드 |
| `CONT_CODE` | 콘텐츠 코드 |
| `CONT_NAME` | 콘텐츠 이름 |
| `META_DATA` | 메타 데이터 |

### `StatisticWord`

| Field | 의미 |
| --- | --- |
| `WORD` | 용어 |
| `CONTENT` | 설명 |

## 주기와 기간 형식

| 주기 | 코드 | 기간 예시 |
| --- | --- | --- |
| 연 | `A` | `2020` |
| 반년 | `S` | `2020S1` |
| 분기 | `Q` | `2020Q1` |
| 월 | `M` | `202001` |
| 반월 | `SM` | `202001S1` |
| 일 | `D` | `20200101` |

## 공통 메시지 코드

| 구분 | 코드 | 의미 |
| --- | --- | --- |
| 정보 | `APIMSG001-100` | 인증키가 유효하지 않음 |
| 정보 | `APIMSG001-200` | 해당 데이터 없음 |
| 에러 | `APIMSG002-100` | 필수 값 누락 |
| 에러 | `APIMSG002-101` | 주기와 다른 날짜 형식 |
| 에러 | `APIMSG002-200` | 파일타입 값 누락 또는 유효하지 않음 |
| 에러 | `APIMSG002-300` | 조회시작/종료건수 누락 |
| 에러 | `APIMSG002-301` | 조회건수 타입 오류 |
| 에러 | `APIMSG002-400` | 검색 범위 초과로 60초 timeout |
| 에러 | `APIMSG002-500` | 서버 오류 또는 서비스 찾을 수 없음 |
| 에러 | `APIMSG002-600` | DB Connection 오류 |
| 에러 | `APIMSG002-601` | SQL 오류 |
| 에러 | `APIMSG002-602` | 과도한 호출로 이용 제한 |

## 추적 기준

- `Service`, `Params`, `Response fields`는 공식 문서에서 확인한 내용만 확정값으로 적는다.
- 블로그나 예제 코드는 공식 문서가 비어 있을 때만 보조 근거로 사용한다.
- SDK typed method 대응표는 `docs/apis/typed-sdk-checklist.md`에서 별도로 추적한다.
- 전체 요청/응답 필드 스키마는 추후 OpenAPI 산출물에서 확인한다.
