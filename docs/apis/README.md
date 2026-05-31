# API 목록

이 디렉터리는 한국은행 ECOS 공식 OpenAPI의 전체 API 목록, 요청 파라미터, 응답 필드, SDK method 대응 관계를 추적한다.

## SDK typed 구현

ECOS 공식 문서 기준 API 목록은 새로 수집한다.

- 전체 대응표: `docs/apis/typed-sdk-checklist.md`
- 기준 인벤토리: `docs/apis/official-inventory.md`
- 조사 노트: `docs/apis/ecos-openapi-research.md`
- OpenAPI 생성 기준: `docs/apis/openapi.md`
- SDK package: `github.com/awuzag/ecos`

## 구현 방식

- root SDK는 `Client`와 `Config`를 수동 코드로 유지한다.
- API별 typed method와 응답 타입은 공식 문서 수집 후 추가한다.
- 생성 흐름을 도입하면 생성물은 직접 수정하지 않는다.
- 사람이 정하는 friendly SDK method 이름은 별도 override 문서로 분리한다.
- 에러와 진단은 인증키를 노출하지 않는다.

## 공통 호출 형식

ECOS OpenAPI는 path segment 기반 호출 형식을 사용한다.

```text
https://ecos.bok.or.kr/api/{service}/{apiKey}/{format}/{lang}/{start}/{end}/...
```

초기 기본값:

- base URL: `https://ecos.bok.or.kr/api`
- format: `json`
- lang: `kr`
- auth env: `ECOS_API_KEY`

## 공통 정리 대상

- service 이름
- 필수 path segment
- 선택 path segment
- paging 기준
- 응답 root key
- provider-native error code/message
- 항목 코드와 통계표 코드의 관계
- 주기별 기간 형식
