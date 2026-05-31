# ECOS OpenAPI 조사 노트

- 확인 날짜: 2026-05-31
- 공식 서비스 소개: https://ecos.bok.or.kr/api/#/ServiceIntroduction
- 서비스 이용방법: https://ecos.bok.or.kr/api/#/ServiceUse/ServiceUseHow
- 서비스 목록: https://ecos.bok.or.kr/api/#/ServiceUse/ServiceList
- 개발 명세서: https://ecos.bok.or.kr/api/#/DevGuide/DevSpeciflcation
- 통계코드검색: https://ecos.bok.or.kr/api/#/DevGuide/StatisticalCodeSearch

## 요약

한국은행 ECOS OpenAPI는 ECOS 경제통계 정보를 외부 프로그램에서 조회할 수 있게 제공하는 HTTP API다. 호출 조건 대부분이 URL path segment에 들어가며, 인증키도 path segment에 포함된다.

일반적인 사용 흐름은 다음과 같다.

1. `StatisticTableList`로 통계표 코드를 찾는다.
2. `StatisticItemList`로 해당 통계표의 항목 코드, 주기, 제공 기간, 단위를 확인한다.
3. `StatisticSearch`로 실제 시계열 값을 조회한다.
4. 대표 최신값만 필요하면 `KeyStatisticList`를 사용한다.
5. 설명 자료나 용어 설명이 필요하면 `StatisticMeta`, `StatisticWord`를 사용한다.

## 서비스 이용 절차

공식 사이트가 안내하는 서비스 이용 절차는 다음과 같다.

1. 한국은행 Open API 사이트 접속
2. Open API 인증키 신청
3. Open API 검색 및 이용방법 확인
4. Open API를 이용해 애플리케이션 제작
5. 애플리케이션 등록

운영 호출에는 발급받은 인증키가 필요하다. 문서와 테스트 fixture에는 실제 인증키를 남기지 않는다.

## 공통 URL 형식

```text
https://ecos.bok.or.kr/api/{service}/{apiKey}/{format}/{lang}/{start}/{end}/{service-specific-segments...}
```

공통 segment:

| Segment | 예시 | 필수 | 설명 |
| --- | --- | --- | --- |
| `service` | `StatisticSearch` | Y | ECOS API 서비스명 |
| `apiKey` | `sample` | Y | 발급받은 인증키. 공식 샘플 URL에서는 `sample` 사용 |
| `format` | `json` | Y | `json`, `xml` |
| `lang` | `kr` | Y | `kr`, `en` |
| `start` | `1` | Y | 요청 시작 건수 |
| `end` | `10` | Y | 요청 종료 건수 |

응답은 JSON 기준으로 대체로 다음 형태를 가진다.

```json
{
  "ServiceName": {
    "list_total_count": 1,
    "row": []
  }
}
```

`KeyStatisticList`는 `row_count`도 함께 내려온다.

## 지원 서비스 목록

공식 서비스 목록에서 확인한 ECOS OpenAPI 서비스는 6개다.

| Service | 한국어 서비스명 | 용도 | 샘플 URL |
| --- | --- | --- | --- |
| `StatisticTableList` | 서비스 통계 목록 | OpenAPI 대상 통계표 목록 제공 | `https://ecos.bok.or.kr/api/StatisticTableList/sample/json/kr/1/10/` |
| `StatisticItemList` | 통계 세부항목 목록 | 특정 통계표의 항목 코드, 주기, 기간, 단위 제공 | `https://ecos.bok.or.kr/api/StatisticItemList/sample/json/kr/1/10/601Y002/` |
| `StatisticSearch` | 통계 조회 조건 설정 | 통계표 코드, 항목 코드, 기간으로 시계열 데이터 조회 | `https://ecos.bok.or.kr/api/StatisticSearch/sample/json/kr/1/10/200Y101/A/2020/2023/10101/?/?/?` |
| `KeyStatisticList` | 100대 통계지표 | 주요 지표 최신값 스냅샷 제공 | `https://ecos.bok.or.kr/api/KeyStatisticList/sample/json/kr/1/10/` |
| `StatisticMeta` | 통계메타DB | 통계 설명자료 제공 | `https://ecos.bok.or.kr/api/StatisticMeta/sample/json/kr/1/10/경제심리지수/` |
| `StatisticWord` | 통계용어사전 | 경제·통계 용어 설명 제공 | `https://ecos.bok.or.kr/api/StatisticWord/sample/json/kr/1/10/소비자동향지수/` |

## 주기와 기간 형식

`StatisticSearch`는 `cycle`, `start_time`, `end_time`의 형식이 맞아야 한다.

| 주기 | 코드 | 기간 예시 | 설명 |
| --- | --- | --- | --- |
| 연 | `A` | `2020` | 연간 |
| 반년 | `S` | `2020S1` | 반기 |
| 분기 | `Q` | `2020Q1` | 분기 |
| 월 | `M` | `202001` | 월 |
| 반월 | `SM` | `202001S1` | 월 중 전반/후반 |
| 일 | `D` | `20200101` | 일 |

주의 사항:

- 주기와 맞지 않는 날짜 형식을 보내면 `APIMSG002-101`이 발생할 수 있다.
- 조회 범위가 너무 넓으면 60초 timeout 성격의 `APIMSG002-400`이 발생할 수 있다.
- 같은 `ITEM_CODE`라도 `A`, `M`, `Q`, `D` 등 주기별로 제공 기간과 데이터 개수가 다를 수 있다.

## 핵심 호출 흐름

### 1. 통계표 찾기

`StatisticTableList`는 통계표 계층과 실제 조회 가능 여부를 제공한다.

```text
GET /StatisticTableList/{apiKey}/json/kr/1/10/
```

예시 응답 필드:

| Field | 의미 |
| --- | --- |
| `P_STAT_CODE` | 상위 통계 코드 |
| `STAT_CODE` | 통계표 코드 |
| `STAT_NAME` | 통계표 이름 |
| `CYCLE` | 대표 주기. 계층 노드이면 `null`일 수 있음 |
| `SRCH_YN` | 실제 조회 가능 여부. `Y`이면 `StatisticSearch` 대상 |
| `ORG_NAME` | 작성 기관 |

예시:

| `STAT_CODE` | `STAT_NAME` | `CYCLE` | `SRCH_YN` |
| --- | --- | --- | --- |
| `102Y001` | 본원통화 구성내역(말잔, 원계열) | `M` | `Y` |
| `722Y001` | 한국은행 기준금리 및 여수신금리 | `D` | `Y` |
| `817Y002` | 시장금리(일별) | `D` | `Y` |

### 2. 통계 세부항목 찾기

`StatisticItemList`는 특정 통계표 안의 항목 코드와 주기별 제공 기간을 제공한다.

```text
GET /StatisticItemList/{apiKey}/json/kr/1/100/{statCode}/
```

예시 응답 필드:

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
| `CYCLE` | 해당 항목의 제공 주기 |
| `START_TIME` | 제공 시작 시점 |
| `END_TIME` | 제공 종료 시점 |
| `DATA_CNT` | 데이터 개수 |
| `UNIT_NAME` | 단위 |
| `WEIGHT` | 가중치 |

예시: `722Y001`의 한국은행 기준금리

| `ITEM_CODE` | `ITEM_NAME` | `CYCLE` | `START_TIME` | `END_TIME` | `UNIT_NAME` |
| --- | --- | --- | --- | --- | --- |
| `0101000` | 한국은행 기준금리 | `D` | `19990506` | `20260529` | `연%` |
| `0101000` | 한국은행 기준금리 | `M` | `199905` | `202604` | `연%` |
| `0101000` | 한국은행 기준금리 | `Q` | `1999Q2` | `2026Q1` | `연%` |
| `0101000` | 한국은행 기준금리 | `A` | `1999` | `2025` | `연%` |

### 3. 실제 시계열 조회

`StatisticSearch`는 실제 값을 조회한다.

```text
GET /StatisticSearch/{apiKey}/json/kr/{start}/{end}/{statCode}/{cycle}/{startTime}/{endTime}/{itemCode1}/{itemCode2}/{itemCode3}/{itemCode4}
```

사용하지 않는 item code segment는 공식 샘플처럼 `?`를 넣는다.

예시: 한국은행 기준금리 월별 조회

```text
GET /StatisticSearch/{apiKey}/json/kr/1/100/722Y001/M/202001/202604/0101000/?/?/?
```

예시 응답 필드:

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

## 100대 통계지표와 과거 데이터

`KeyStatisticList`는 주요 지표의 최신값 스냅샷에 가깝다.

```text
GET /KeyStatisticList/{apiKey}/json/kr/1/10/
```

예시 필드:

| Field | 의미 |
| --- | --- |
| `CLASS_NAME` | 지표 분류 |
| `KEYSTAT_NAME` | 지표명 |
| `DATA_VALUE` | 최신값 |
| `CYCLE` | 이 API에서는 주기 코드가 아니라 값의 기준시점 |
| `UNIT_NAME` | 단위 |

예시:

```json
{
  "CLASS_NAME": "시장금리",
  "KEYSTAT_NAME": "한국은행 기준금리",
  "DATA_VALUE": "2.5",
  "CYCLE": "20260528",
  "UNIT_NAME": "%"
}
```

위 값은 `2026-05-28` 기준 한국은행 기준금리가 `2.5%`라는 뜻이다. 과거 데이터는 `KeyStatisticList`가 아니라 같은 지표에 대응되는 `StatisticSearch` 호출로 조회한다.

대표 매핑 예시:

| 지표 | `STAT_CODE` | `ITEM_CODE` | 비고 |
| --- | --- | --- | --- |
| 한국은행 기준금리 | `722Y001` | `0101000` | 일/월/분기/연 제공 |
| KORIBOR(3개월) | `817Y002` | `010150000` | 일별 시장금리 |
| CD(91일) | `817Y002` | `010502000` | 일별 시장금리 |
| 통안증권(1년) | `817Y002` | `010400001` | 일별 시장금리 |

## 에러 코드

공식 개발 명세에서 확인한 공통 메시지 코드:

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

## 구현 전 메모

- `StatisticTableList`의 `SRCH_YN=Y`인 행만 실제 조회 대상 통계표로 본다.
- `StatisticTableList.CYCLE`은 대표 주기이고, 실제 지원 주기는 `StatisticItemList`의 항목별 `CYCLE`을 우선한다.
- `KeyStatisticList.CYCLE`은 최신값의 기준시점으로 해석한다.
- 기준금리처럼 정책성 지표는 대부분의 장기 분석에서 월별 `M` 조회만으로 충분하다.
- 일별 `D`는 이벤트 분석, 날짜 정합성이 중요한 백테스트, 시장 데이터와의 일별 조인에 필요하다.
- 한글 path segment가 들어가는 `StatisticMeta`, `StatisticWord`는 URL escaping이 필요하다.
- 실제 인증키는 URL path에 들어가므로 error, log, fixture, 문서 예시에 노출하지 않는다.
