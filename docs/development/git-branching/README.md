# Git Branching Strategy

## 목적

이 문서는 `ecos` 의 브랜치 운영 기준을 정리한다.

`ecos` 는 Go SDK이고, 이후 CLI를 제공할 수 있다. 사용자는 최종적으로 특정 버전을 설치하거나 의존성으로 고정한다. 따라서 작업 브랜치와 배포 기준을 분리한다. 작업 브랜치는 목적별 접두사를 사용하고, 배포 기준은 릴리스 브랜치와 SemVer 태그로 고정한다.

## 기본 방향

- `main` 은 항상 다음 릴리스 후보가 모이는 기준선이다.
- `feat/*`, `fix/*` 는 원격에 push 할 수 있는 작업 브랜치다.
- 리모트에 만들 수 있는 브랜치는 `main`, `release/*`, `feat/*`, `fix/*` 로 제한한다.
- `codex/*`, `claude/*` 처럼 도구나 워크트리 출처를 드러내는 접두사는 리모트 브랜치 이름으로 쓰지 않는다.
- 사용자가 설치하는 기준은 브랜치가 아니라 `vX.Y.Z` 태그다.

## 브랜치 종류

| 브랜치 | 위치 | 용도 |
| --- | --- | --- |
| `main` | local, remote | 검증된 변경이 모이는 기본 브랜치 |
| `feat/<topic>` | local, remote | 작은 기능 구현 |
| `fix/<topic>` | local, remote | 버그 수정 또는 설정 수정 |
| `release/vX.Y` | local, remote | `vX.Y.Z` 패치 릴리스 안정화 |
| `vX.Y.Z` tag | remote | 사용자가 설치하는 고정 버전 |

`feat/*`, `fix/*` 는 원격에 push 할 수 있지만 배포 기준은 아니다. PR, CI, 리뷰, 작업 공유를 위한 브랜치로 사용하고, 검증된 변경만 `main` 또는 `release/*` 로 통합한다.

## 작업 흐름

일반 기능 작업은 작은 로컬 브랜치에서 시작한다.

```bash
git switch main
git pull --ff-only
git switch -c feat/statistic-search
git push -u origin feat/statistic-search
```

작업이 끝나면 로컬에서 검증하고 PR 또는 명시적인 merge 절차로 `main` 에 통합한다.

```bash
go test ./...
git switch main
git merge --ff-only feat/statistic-search
```

## 배포 흐름

SDK/CLI 배포는 `release/*` 브랜치와 SemVer 태그를 함께 사용한다.

```bash
git switch main
git pull --ff-only
git switch -c release/v0.1
git push -u origin release/v0.1
```

배포 준비가 끝나면 태그를 만든다.

```bash
git tag v0.1.0
git push origin v0.1.0
```

사용자 설치 기준은 태그다.

```bash
go get github.com/awuzag/ecos@v0.1.0
```

## 금지 규칙

- `codex/*`, `claude/*`, `worktree/*` 같은 도구별 접두사를 `origin` 에 push 하지 않는다.
- 허용 목록 밖의 top-level 브랜치를 만들지 않는다.
- `dev`, `staging`, `prod` 같은 장기 환경 브랜치를 만들지 않는다.
- 배포 기준을 움직이는 브랜치 이름으로 안내하지 않는다.
- 검증하지 않은 실험 브랜치에서 태그를 만들지 않는다.
- 여러 주제의 변경을 한 브랜치와 한 커밋에 섞지 않는다.

## 관련 결정

- `docs/adr/0001-sdk-release-branching-strategy.md`
