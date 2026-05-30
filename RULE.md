# RULE.md

이 저장소는 `github.com/awuzag/ecos` 독립 Go SDK를 제공한다. `mwosa`나 다른 상위 애플리케이션에 의존하지 않고, 한국은행 경제통계시스템(ECOS) OpenAPI를 얇고 안정적인 타입 메서드로 감싼다.

## Go 코드

- package name은 `ecos`를 사용한다.
- 공개 생성자는 `New(Config, ...Option)` 형태를 유지한다.
- 필수 값은 명시적인 `Config` struct에 둔다.
- 선택 값은 `With...` option으로 추가한다.
- 내부 HTTP 호출은 `resty`를 사용한다.
- HTTP client 직접 구현보다 `github.com/go-resty/resty`의 client, request, response API를 우선 사용한다.
- I/O를 수행하는 공개 메서드는 첫 인자로 `context.Context`를 받는다.
- 공개 타입과 메서드는 Go doc comment를 작성한다.
- 불필요한 `any`, `interface{}`, broad catch, silent fallback을 피한다.
- 자동 생성 코드는 이 규칙의 수동 코드 제한에서 제외하되, 생성물은 직접 수정하지 않는다.

## Error Handling

- 직접 작성하는 Go 코드는 error 생성, wrapping, joining에 `github.com/samber/oops`를 사용한다.
- 새 error를 만들 때 `fmt.Errorf`, `errors.New`, `errors.Join`을 사용하지 않는다.
- error 판별을 위한 `errors.Is`, `errors.As`는 사용할 수 있다.
- 새 error는 `oops.New` 또는 `oops.Errorf`를 사용한다.
- 하위 레이어 error를 원인으로 보존해야 할 때는 `Wrap` 또는 `Wrapf`를 사용한다.
- cleanup 과정에서 여러 error를 보존해야 할 때는 `oops.Join`을 사용한다.
- 같은 함수 안에서 domain과 context가 반복되면 재사용 가능한 builder를 먼저 만든다.
- builder는 `.New`, `.Errorf`, `.Wrap`, `.Wrapf`, `.Join`, `.Recover`, `.Recoverf` 같은 종료 메서드로 끝낸다.
- HTTP status error, JSON decode error, ECOS business error를 구분한다.
- ECOS 응답의 provider-native error code/message는 손실 없이 보존한다.
- decode 실패를 성공처럼 처리하지 않는다.
- secret, 인증키, 원문 credential은 error string, log, test fixture에 남기지 않는다.

예시:

```go
errb := oops.
	In("statistic_search").
	With(
		"service", "StatisticSearch",
		"stat_code", params.StatCode,
		"cycle", params.Cycle,
	)

resp, err := client.resty.R().
	SetContext(ctx).
	Get(client.buildPath("StatisticSearch", pathSegments...))
if err != nil {
	return Result{}, errb.Wrap(err)
}
```

## Tests

- 기본 테스트는 offline fake HTTP server 기반으로 작성한다.
- 단위 테스트에서는 `testing`과 `github.com/stretchr/testify`를 적극적으로 사용한다.
- 실패 시 이후 검증이 의미 없으면 `require`를 우선하고, 같은 상태에서 여러 값을 함께 확인할 때는 `assert`를 사용할 수 있다.
- 테스트 helper는 실패를 숨기지 않는다.
- helper 안에서 테스트를 중단해야 한다면 `t.Helper()`와 `require`로 실패 위치를 호출자 기준으로 드러낸다.
- 기본 `go test ./...`는 빠르고 재현 가능한 단위 테스트와 가벼운 통합 테스트를 대상으로 한다.
- 실제 외부 API 호출이나 사용자의 로컬 환경에 강하게 묶인 테스트는 기본 경로에 넣지 않는다.
- 단위 테스트는 함수, 메서드, 작은 패키지 단위를 검증한다.
- 외부 의존성은 fake HTTP transport 또는 `httptest.Server`로 대체한다.
- 빌드된 CLI 실행, 외부 프로세스, 실제 provider API처럼 느리거나 환경 의존성이 큰 검증은 `integration` 또는 `e2e` build tag로 분리한다.
- live ECOS 호출은 기본 `go test ./...`에 포함하지 않는다.
- live/e2e가 필요하면 `//go:build e2e`와 secret 기반 optional workflow로 분리한다.

## Documentation

- 공식 ECOS 문서에서 확인한 요청 파라미터, 응답 필드, 메시지 코드를 repo-local 문서로 남긴다.
- 링크 목록만 남기지 않고 SDK 구현에 필요한 필드 의미와 제약을 같이 정리한다.
- 문서 갱신 시 기준 URL과 확인 날짜를 함께 남긴다.
- 통계표 코드, 주기, 기간 형식, 항목 코드의 차이를 README와 API 문서에 명확히 적는다.

## Secrets

- 인증키, 샘플 secret, 개인 계정 값은 커밋하지 않는다.
- README와 테스트는 `ECOS_API_KEY` 같은 환경변수 이름만 예시로 사용한다.
- fixture에는 실제 ECOS 인증키나 실사용 계정 정보를 넣지 않는다.

## Verification

기본 변경은 아래 명령을 통과해야 한다.

```sh
go mod tidy
go test ./...
git diff --check
```

가능하면 아래 검증도 함께 수행한다.

```sh
go test -race ./...
go vet ./...
```
