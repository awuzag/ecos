# ECOS ROI-first catalog

- 확인 날짜: 2026-06-07 KST
- 기준 API: `KeyStatisticList`, `StatisticSearch`, `StatisticTableList`, targeted `StatisticItemList`
- key statistics schema: `catalog/schema/key-statistics-snapshot.schema.json`
- inventory schema: `catalog/schema/inventory-snapshot.schema.json`
- logical indicator schema: `catalog/schema/series-aliases.schema.json`

## 결론

ECOS catalog의 1순위 목적은 전체 provider surface를 모두 펼치는 것이 아니라 ROI 높은 거시/시장 지표를 안정적으로 찾고 조회하는 것이다.

기본 entry point는 `KeyStatisticList` 100대 통계지표다. 이 API는 주요 지표의 최신값, 기준시점, 단위를 빠르게 보여주는 상태판으로 취급한다. 특정 지표의 과거 흐름이 필요하면 catalog의 logical indicator mapping을 통해 `StatisticSearch` 호출로 drill-down한다.

원본 API CLI는 계속 escape hatch로 유지한다.

```text
ecos show key-statistics
ecos list tables
ecos list items --stat-code 722Y001
ecos search observations --stat-code 722Y001 --cycle M --time-start 202001 --time-end 202604 --item-code 0101000
```

## 역할 분리

- `KeyStatisticList`: 100대 주요 지표의 최신 상태판. `CYCLE` provider 필드는 SDK에서 `ReferenceTime`으로 노출하며 일반 cycle enum이 아니라 최신값의 기준시점이다.
- `catalog/aliases/series.json`: 사람이 쓰는 논리 지표 catalog. 예를 들어 `bok-base-rate` 하나가 `D/M/Q/A` variant를 가진다.
- `StatisticSearch`: 관심 지표의 과거 시계열 조회 경로. `stat_code`, `cycle`, `item_codes`, 기간을 명시한다.
- `StatisticTableList`: 전체 provider surface와 drift 감지를 위한 얕은 table snapshot.
- targeted `StatisticItemList`: catalog에 올릴 지표 또는 drift 확인 대상만 항목/기간/단위까지 확인한다.
- full `StatisticItemList` deep crawl: 기본 경로가 아니다. 필요할 때만 낮은 동시성, 제한된 target, 재개 옵션으로 실험한다.

## catalog 구조

```text
catalog/
  key-statistics/
    README.md
  inventory/
    README.md
    tables.json             # optional generated shallow table snapshot
    targeted-items.json     # optional generated targeted item snapshot
  aliases/
    series.json             # human-maintained logical indicator aliases
  schema/
    key-statistics-snapshot.schema.json
    inventory-snapshot.schema.json
    series-aliases.schema.json
```

역할:

- `catalog/key-statistics/`: 100대 통계지표 snapshot의 위치와 갱신 규칙.
- `catalog/inventory/`: full deep crawl 전제가 아니라 shallow table snapshot과 targeted item snapshot의 위치.
- `catalog/aliases/series.json`: high-ROI logical indicator catalog.
- `catalog/schema/*.schema.json`: snapshot과 alias의 machine-checkable 계약.

## logical indicator 모델

Cycle별 series를 별도 alias로 쪼개지 않고 하나의 경제 지표 아래 묶는다.

```json
{
  "id": "bok-base-rate",
  "stat_code": "722Y001",
  "item_codes": ["0101000"],
  "preferred_cycle": "M",
  "available_cycles": ["D", "M", "Q", "A"],
  "periods": [
    {
      "cycle": "M",
      "series_key": "722Y001:M:0101000",
      "start_time": "199905",
      "end_time": "202605",
      "data_count": 325,
      "unit_name": "연%",
      "status": "queryable"
    }
  ]
}
```

규칙:

- alias `id`는 `bok-base-rate-monthly`보다 `bok-base-rate`처럼 논리 지표명 중심으로 만든다.
- `preferred_cycle`은 기본 drill-down 주기다. 사용자가 명시하면 다른 `available_cycles`를 조회할 수 있다.
- `periods`는 cycle별 `series_key`, 제공 기간, 데이터 수, 단위, 검증 상태를 담는다.
- `series_key`는 `STAT_CODE:CYCLE:ITEM_CODE1[:ITEM_CODE2...]` 형식으로 `StatisticSearch` mapping을 드러낸다.
- `KeyStatisticList`의 최신값과 `StatisticSearch`의 과거 series는 이름이 비슷해도 정의가 다를 수 있다. 이 경우 `needs_review`로 둔다.

## collector 방향

`scripts/collect-series-catalog-inventory`는 세 층으로 쓴다.

1. 기본 수집: `StatisticTableList` + `KeyStatisticList`.
2. 관심 지표 수집: `-include-items -stat-code ...`로 targeted `StatisticItemList`만 조회.
3. 실험적 deep crawl: `-allow-full-item-crawl`을 명시하고 낮은 동시성, sleep, limit, resume를 함께 둔다.

권장 실행:

```sh
go run ./scripts/collect-series-catalog-inventory -out tmp/ecos-inventory-smoke.json
go run ./scripts/collect-series-catalog-inventory -include-items -stat-code 722Y001 -out tmp/ecos-inventory-722Y001.json
go run ./scripts/collect-series-catalog-inventory -include-items -stat-code 722Y001 -stat-code 817Y002 -concurrency 1 -sleep 500ms -out tmp/ecos-inventory-targeted-rates.json
```

`-include-items`는 전체 searchable table deep crawl 권장이 아니다. 기본적으로 하나 이상의 `-stat-code`가 필요하다. target 없이 전체를 순회하려면 `-allow-full-item-crawl`을 명시해야 하며, 이 출력은 `experimental_full_items` scope의 실험 산출물로만 취급한다.

실험적 deep crawl을 할 때는 다음을 기본값으로 삼는다.

- `-concurrency 1`
- `-sleep 500ms` 이상
- `-max-item-tables` 또는 반복 `-stat-code`로 범위 제한
- 실패 후 `-start-stat-code`로 재개
- 출력은 repo-local `tmp/...` 파일

```sh
go run ./scripts/collect-series-catalog-inventory -include-items -allow-full-item-crawl -concurrency 1 -sleep 500ms -max-item-tables 10 -out tmp/ecos-inventory-experimental-items.json
```

inventory snapshot의 `source.scope`는 다음처럼 해석한다.

- `tables_and_key_statistics`: 얕은 table snapshot과 100대 통계지표 상태판.
- `targeted_items`: 관심 통계표만 `StatisticItemList`로 확인한 drill-down 검증 산출물.
- `experimental_full_items`: 전체 또는 넓은 범위 item crawl 실험 산출물. source of truth가 아니다.

스크립트는 `--api-key`, `ECOS_API_KEY`, repo-local `.env` 순서로 키를 읽고 값은 출력하지 않는다. 자동 retry/backoff는 SDK 기본 동작에 넣지 않는다. `APIMSG002-602` 같은 제한은 collector 옵션과 운영 절차로 명시적으로 다룬다.

## 다중 group table

다중 group table은 자동 Cartesian product로 전부 펼치지 않는다. 예를 들어 환율처럼 `ITEM_CODE1`, `ITEM_CODE2` 조합이 필요한 table은 다음 중 하나로 처리한다.

- 사람이 확인한 target mapping을 `catalog/aliases/series.json`에 `needs_review` 또는 `queryable`로 기록한다.
- targeted `StatisticSearch` smoke로 실제 조회를 확인한 뒤 `periods`를 채운다.
- 전체 조합 생성을 drift/실험 산출물로만 남기고 기본 catalog에는 승격하지 않는다.

## tmp 실험 산출물

`tmp/ecos-inventory-full.json`은 source of truth가 아니다. 이전 실험에서 searchable table 119개, item 5,905개, series 후보 12,347개 partial snapshot까지 생성됐고, 병렬 deep crawl은 ECOS `APIMSG002-602` rate limit에 걸렸다.

이 파일은 다음 목적에만 쓴다.

- full crawl이 현재 제품 목적 대비 노이즈가 크다는 반례.
- 일부 지표 mapping을 다시 확인하기 위한 참고 자료.
- collector durability와 resume 동작 확인용 fixture 성격의 실험 산출물.

## drift detection

drift detection은 전체 item deep crawl보다 얕은 table snapshot과 targeted indicator snapshot을 우선 비교한다.

변경 유형:

- `added_table`, `removed_table`
- `changed_table_name`, `changed_searchable`, `changed_table_cycle`
- `changed_indicator_period`
- `changed_indicator_unit`
- `changed_indicator_mapping`
- `changed_key_statistic_name`

위험도 기준:

- 낮음: `end_time`, `data_count` 증가처럼 정상 갱신 가능성이 큰 period 변경.
- 중간: 이름 표기 변경, key statistic class 변경.
- 높음: `unit_name`, `stat_code`, `item_codes`, `available_cycles`, `searchable` 변경.

## Go 생성 방향

생성 입력 후보:

```text
catalog/aliases/series.json
catalog/inventory/*.json
catalog/key-statistics/*.json
```

생성 후보:

- `IndicatorID` typed const.
- `LookupIndicator(id)` helper.
- `Indicator.SearchRequest(cycle, start, end)` helper.
- 필요한 경우 `StatCode` typed const.

규칙:

- generated file은 직접 수정하지 않는다.
- 수동 SDK core는 `Client`, `Config`, request helper, error type 중심으로 작게 유지한다.
- 원본 API 접근 명령은 catalog UX와 별개로 유지한다.
- `needs_review` mapping은 typed convenience helper로 승격하지 않는다.

## 검증

JSON 문법:

```sh
python3 -m json.tool catalog/aliases/series.json >/dev/null
python3 -m json.tool catalog/schema/key-statistics-snapshot.schema.json >/dev/null
python3 -m json.tool catalog/schema/inventory-snapshot.schema.json >/dev/null
python3 -m json.tool catalog/schema/series-aliases.schema.json >/dev/null
```

JSON Schema:

```sh
npx --yes ajv-cli@5 validate -s catalog/schema/series-aliases.schema.json -d catalog/aliases/series.json --strict=false
npx --yes ajv-cli@5 validate -s catalog/schema/inventory-snapshot.schema.json -d tmp/ecos-inventory-smoke.json --strict=false
```

CLI smoke:

```sh
go run ./cmd/ecos show key-statistics --start 1 --end 10 --json
go run ./cmd/ecos list items --stat-code 722Y001 --start 1 --end 10 --json
```

기본 repo 검증:

```sh
go mod tidy
go test ./...
git diff --check
```
