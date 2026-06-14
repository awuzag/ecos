# ECOS catalog

`catalog/`는 100대 통계지표 기반 상태판과 관심 지표 drill-down mapping을 관리한다.

```text
catalog/
  key-statistics/
    README.md
  inventory/
    README.md
    tables.json             # optional generated shallow table snapshot
    targeted-items.json     # optional generated targeted item snapshot
  aliases/
    series.json             # logical indicator aliases
  schema/
    key-statistics-snapshot.schema.json
    inventory-snapshot.schema.json
    series-aliases.schema.json
```

## Source of truth

- `KeyStatisticList`는 catalog의 최신값 entry point다.
- `catalog/aliases/series.json`은 high-ROI logical indicator layer다. 사람이 쓰는 ID와 `StatisticSearch` mapping을 연결한다.
- `catalog/inventory/`는 shallow table snapshot과 targeted item snapshot을 둔다. 전체 `StatisticItemList` deep crawl을 기본 전제로 삼지 않는다.
- `tmp/ecos-inventory-full.json` 같은 실험 산출물은 source of truth가 아니다.

Inventory snapshot은 `source.scope`로 성격을 구분한다. 기본값은 `tables_and_key_statistics`이고, 관심 통계표 검증은 `targeted_items`, 넓은 item crawl 실험은 `experimental_full_items`다.

## Key format

`series_key`는 `StatisticSearch` 호출에 필요한 provider mapping을 담는다.

```text
STAT_CODE:CYCLE:ITEM_CODE1[:ITEM_CODE2][:ITEM_CODE3][:ITEM_CODE4]
```

예: `722Y001:M:0101000`.

Alias ID는 cycle별 series ID가 아니라 논리 지표명이다. 예: `bok-base-rate`.
