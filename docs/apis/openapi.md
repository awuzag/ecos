# OpenAPI 수집 계획

- 확인 날짜: 미수집
- 루트 산출물: `docs/apis/ecos.openapi.json`
- 번들 산출물: `docs/apis/ecos.openapi.bundle.json`
- API별 산출물: `docs/apis/openapi/apis/*.json`
- 원본 표 덤프: `docs/apis/ecos-api-metadata.json`
- 재생성 명령: 미정

## 수집 범위

한국은행 ECOS OpenAPI 공식 문서에서 제공하는 전체 service를 수집한다. 수집 전에는 코드 생성물이나 OpenDART/mwosa 전용 API 산출물을 복사하지 않는다.

초기 조사 대상:

- `StatisticTableList`
- `StatisticWord`
- `StatisticItemList`
- `StatisticSearch`
- `KeyStatisticList`
- `StatisticMeta`

## 스키마화 규칙

- 루트 OpenAPI 파일은 전체 API path index로 둔다.
- 번들 OpenAPI 파일은 코드 생성기나 단일 파일 소비 도구를 위한 전체 문서로 둔다.
- 각 API 파일은 API 1개 단위의 operation, parameter, response schema를 함께 가진다.
- 인증키는 path segment에 들어가므로 OpenAPI security scheme이나 path parameter로 일관되게 표현한다.
- 공식 요청 타입과 길이 제약이 확인되면 schema에 반영한다.
- JSON 응답의 공통 error 구조와 service별 data row 구조를 분리한다.
- provider-native 필드명은 손실 없이 보존하고, SDK friendly type은 별도 mapping으로 둔다.

## 파일 배치

| 파일 | 용도 |
| --- | --- |
| `docs/apis/ecos.openapi.json` | 전체 API path index |
| `docs/apis/ecos.openapi.bundle.json` | 코드 생성기용 전체 번들 OpenAPI 문서 |
| `docs/apis/openapi/apis/{service}.json` | API 1개 단위 operation, parameter, response schema |
| `docs/apis/ecos-api-metadata.json` | 공식 문서 표를 손실 없이 보존한 원본 파싱 결과 |

## 구현 참고

수집이 끝나면 다음 순서로 SDK에 반영한다.

1. `docs/apis/official-inventory.md`에 공식 service 목록과 파라미터를 기록한다.
2. OpenAPI split 문서를 생성한다.
3. typed request/response type 생성을 검토한다.
4. root SDK의 public method 이름을 `docs/apis/typed-sdk-checklist.md`에서 추적한다.
5. live 호출은 e2e build tag와 `ECOS_API_KEY`가 있을 때만 실행한다.
