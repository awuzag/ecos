# ECOS key statistics snapshots

`KeyStatisticList`는 100대 통계지표의 최신값 상태판이다.

이 디렉터리에는 필요할 때 generated snapshot을 둔다. snapshot은 `catalog/schema/key-statistics-snapshot.schema.json`을 따른다.

원칙:

- `CYCLE` provider 필드는 일반 cycle enum이 아니라 최신값 기준시점이다.
- 최신값 확인은 `KeyStatisticList`에서 시작한다.
- 과거 시계열 분석은 `catalog/aliases/series.json`의 logical indicator mapping을 통해 `StatisticSearch`로 drill-down한다.
- snapshot에는 인증키, sample secret, 개인 값을 남기지 않는다.
