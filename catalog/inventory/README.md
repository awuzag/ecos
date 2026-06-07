# ECOS inventory snapshots

이 디렉터리는 generated snapshot을 위한 자리다. 기본 목적은 전체 provider surface deep copy가 아니라 drift 감지와 targeted drill-down 검증이다.

권장 파일:

- `tables.json`: `StatisticTableList` 기반 shallow table snapshot.
- `targeted-items.json`: 관심 지표의 `StatisticItemList` targeted snapshot.
- `targeted-series.json`: 검증된 관심 지표의 `StatisticSearch` mapping 후보.

`scripts/collect-series-catalog-inventory` 기본 실행은 `StatisticTableList`와 `KeyStatisticList`를 수집한다. 항목 수집은 `-include-items -stat-code ...`처럼 대상 통계표를 명시하는 방식이 기본이다.

전체 searchable table을 대상으로 하는 `StatisticItemList` deep crawl은 rate limit과 노이즈가 크므로 기본 산출물로 취급하지 않는다.
