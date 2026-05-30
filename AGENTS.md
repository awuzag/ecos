# AGENTS.md

이 문서는 Codex, Claude 등 자동화 에이전트가 이 저장소에서 작업할 때 따라야 하는 기본 작업 규칙을 정리한다.

## 기본 작업 규칙

- 이 저장소의 기본 프로그래밍 규칙은 `RULE.md`를 따른다.
- 구현, 리팩터링, 테스트, 문서화 판단을 시작하기 전에 `RULE.md`를 확인한다.
- 이 문서와 `RULE.md`가 충돌하면 프로그래밍 세부 규칙은 `RULE.md`를 우선한다.
- 기존 변경사항을 되돌리지 않는다. 특히 사용자가 만든 미커밋 변경은 명시적인 요청 없이 수정하거나 삭제하지 않는다.
- 작업 전후로 변경 범위를 확인하고, 요청받은 범위를 벗어난 리팩터링은 피한다.
- 작업 전에 현재 브랜치와 원격 추적 상태를 확인한다.

```bash
git status --short --branch
```

## 기술스택

- 언어와 모듈: Go 1.26.3, `github.com/awuzag/ecos`.
- Go 버전은 `go.mod`의 `go` directive와 로컬 검증 환경에서 항시 최신 안정 버전으로 유지한다.
- 제공 형태: root package Go SDK를 우선 제공하고, CLI는 필요가 확인되면 별도 `cmd/ecos`로 추가한다.
- HTTP client: `github.com/go-resty/resty/v2`를 기본으로 사용한다.
- 응답 처리: 한국은행 ECOS OpenAPI JSON 응답을 Go 표준 라이브러리 중심으로 파싱한다.
- 테스트: `testing`, `httptest`, `github.com/stretchr/testify`를 사용한다.
- 문서: Markdown 문서를 `README.md`와 `docs/apis/` 아래에 둔다.
- 인증키: `ECOS_API_KEY` 환경변수를 기본 입력으로 사용하고, secret은 커밋하지 않는다.

## API 문서와 자동 생성

ECOS API 목록과 응답 필드는 새로 수집한다. 생성 코드나 수집 산출물은 원본 저장소에서 복사하지 않는다.

생성 흐름을 도입할 때는 다음 원칙을 따른다.

- 공식 문서에서 확인한 API 목록, 요청 파라미터, 응답 필드를 `docs/apis/` 아래에 먼저 기록한다.
- 생성된 Go 파일은 직접 수정하지 않고, 사람이 정하는 SDK method 이름이나 CLI resource 이름은 별도 override 문서에 둔다.
- root SDK의 수동 코어(`Client`, `Config`, error type, request helper)는 작고 안정적으로 유지한다.

## Git branch strategy

핵심 규칙:

- `codex/*`, `claude/*`, `worktree/*` 같은 도구별 접두사는 리모트에 push 하지 않는다.
- 리모트에 push 할 수 있는 브랜치는 `main`, `release/*`, `feat/*`, `fix/*` 이다.
- 사용자가 설치하는 SDK/CLI 기준은 `vX.Y.Z` SemVer 태그다.
- 배포 안정화는 `release/vX.Y` 브랜치에서 진행한다.
- 일반 기능과 수정 작업은 작은 `feat/*`, `fix/*` 브랜치에서 시작한다.
- GitHub ruleset 으로 허용된 브랜치 이름만 원격에 생성한다.

작업 브랜치에서 작업한 변경은 검증 후 PR 또는 명시적인 merge 절차로 `main` 또는 필요한 `release/*` 브랜치에 통합한다.
