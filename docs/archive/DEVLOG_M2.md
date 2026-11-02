# 개발 일지 — M2 · 인증

> **M2 기간의 DEVLOG 날짜 블록**을 옮겨 둔 문서다. 보관만 하고 고치지 않는다(링크 경로만 이 폴더에 맞췄다).
> 지금 일지는 [../DEVLOG.md](../DEVLOG.md), 단계와 완료 기준은 [../PLAN.md](../PLAN.md).

---

## 2026-10-30 · M1 · M2 집 PC 확인 (집 PC)

### 한 일
- 집 PC 준비: winget으로 Go 1.27.0 · pnpm 12.15.0 · Task 3.54.0 · golangci-lint 2.15.0(회사 PC · CI와 같은 버전) 설치, `.env`, `docker compose up -d`, `task setup` · `task migrate`(00001 · 00002 적용), `task user -- create`로 내 계정 (사용자 직접).
- 확인(관찰값):
  - `task test`: Go 패키지 전부 ok — 통합 테스트가 컨테이너에서 실제로 돌았다(`internal/db` 6.1초 · `internal/server` 8.7초. 건너뛰면 1초 안에 끝난다). Vitest 9 통과.
  - `task lint`: golangci-lint 0 issues, svelte-check 0 errors, Prettier · ESLint 통과.
  - `task dev`: 화면 Postgres · Valkey ok, 로그인 뒤 인사말 → **M1 완료 기준 충족**.
  - 로그인 유지 해제로 로그인 → 개발자 도구 쿠키 `session`의 Expires = "Session" → Chrome을 닫았다 열면 로그인 화면. 로그아웃 → `POST /api/auth/logout` 204 → 쿠키 즉시 사라짐 → 다음 `/api/me` 401, 새 탭도 로그인 화면 → **M2 완료 기준 충족**.
- 사용자 질문: "로그인 유지 체크가 다음 로그인 때 기억 안 되는 건 의도인가" → 의도(DESIGN 4.8 기본 해제). 체크하고 로그인한 기기는 30일 동안 로그인 화면을 거의 안 보니 기억해 줘도 편해지는 경우가 드물고, 기본값이 체크로 바뀌면 막으려던 "회사 PC에서 실수로 유지"가 다시 생긴다 → 그대로 둔다.

### 막혔던 일
**도구를 설치했는데 `task`를 못 찾는다**
- 증상: `task setup` → "'task' 용어가 cmdlet … 인식되지 않습니다".
- 원인: 터미널(과 그 터미널을 띄운 VS Code)이 winget 설치 **전에** 열려 있어서 새 PATH를 모른다. `winget list`로는 네 개 다 설치돼 있었다.
- 해결: VS Code를 완전히 껐다 켠다.
- 배운 것: Windows에서 PATH는 프로그램이 시작될 때 한 번 받는다. VS Code 안 터미널은 VS Code의 환경을 물려받아서 터미널 탭만 새로 열어서는 안 될 수 있다.

**"로그인 유지"를 안 했는데 브라우저를 닫았다 열어도 로그인돼 있다**
- 증상: 체크 해제로 로그인 → Chrome 창을 닫았다 열면 이전 페이지가 로그인된 채로 뜬다. 시작 설정은 "새 탭 페이지 열기"였다.
- 원인: Chrome "종료 후에도 백그라운드 앱 계속 실행"이 켜져 있어 창을 닫아도 Chrome 프로세스가 살아 있었다. 세션 쿠키는 브라우저 프로세스가 끝나야 지워진다. 앱은 정상 — 쿠키 Expires가 "Session"(Max-Age 없음)인 것을 개발자 도구로 확인.
- 해결: 그 설정을 끄니 정상. README "자주 막히는 것"에 추가.
- 배운 것: 원인을 브라우저 · 앱으로 가르려면 **결과(로그인 유지됨)가 아니라 중간 증거(쿠키 속성)**를 먼저 본다. 쿠키가 "Session"이면 앱은 할 일을 다 한 것이다. 휴대폰 · 회사 PC에서도 같은 설정이 있으면 같은 일이 생긴다.

### 이번 세션 커밋
(이 커밋) M1 · M2 완료 처리

### 다음에 할 일
- M3 첫 배포: 개념 설명(서버 · 도메인 · HTTPS · 리버스 프록시) → 배포 환경 후보 비교 → 결정 기록 0007 (사용자 결정)

---

## 2026-10-30 · M2 인증 ─ 자동 작업 (회사 PC)

### 한 일
- 사용자 지시로 M2 전체를 자동 작업(의존성만 사용자가 설치). 시작 전에 설계 · 사전 검증 · 진행 순서를 보여 주고 시작했다. 사전 검증의 목표("더 나은 방향으로 완성도 있게, 꼭 무언가를 찾아낼 필요는 없음")를 CLAUDE.md에 적었다.
- 설계 문서를 먼저 고친 뒤 구현: 마이그레이션 → 인증 핵심(argon2id · 입력 규칙 · 유출 확인 · 토큰) → 설정 · 실제 IP → 인증 API · 미들웨어 · 기본 잠금 → 관리 명령 → 프론트 → CI.
- 확인(관찰값):
  - Go 테스트: auth 단위 9 · clientip 7케이스 · config 5 · db 통합 2 · server 통합 16 + 기존 2 — 전부 PASS (실제 Postgres 18 · Valkey 9 컨테이너, HTTPS + 쿠키 저장소). 비밀번호 변경의 캐시 삭제를 일부러 빼면 `TestChangePasswordRevokesOtherSessions`가 200으로 실패하는 것도 확인.
  - 프론트: `pnpm check` 0 · lint 통과 · Vitest 9 · build 성공.
  - 브라우저(임시 계정, 확인 뒤 삭제): `/settings` → `/login?next=%2Fsettings`, 틀린 비밀번호 문구 그대로 표시(로그인 화면에서 안 튕김), 로그인 → `/settings`로 복귀, 닉네임 저장 토스트 "저장했어", `document.cookie`에 세션 없음(HttpOnly), 모바일 하단 탭, 로그아웃 뒤 `/api/me` 401, `?next=/\evil.com`이면 `/`.
  - 관리 명령: 임시 계정으로 create · 짧은 비밀번호 거절 · reset-password · `task user` 경유 create (모두 확인 뒤 삭제).
  - `task gen` 뒤 생성 코드 변화 없음.
  - GitHub Actions(run 41014064520): Go(lint · 통합 테스트 포함 test) · 프론트(check · lint · Vitest) · 생성 코드 확인 · 빌드 세 작업 모두 success — 통합 테스트가 Linux 러너의 Docker에서도 돈다.
- **확인 못 한 것**: 실제 브라우저를 닫았다 여는 "로그인 유지" 동작(쿠키 Max-Age 있음/없음은 통합 테스트로 확인). 집 PC에서 한다.
- 이번 단계는 자동 작업이라 PLAN에 적힌 "개념 설명부터"(유출 확인 · sqlc 등)를 따로 하지 않았다. 코드 주석에 비유와 이유를 적어 두었다.

### [설계] 사전 검증에서 보완한 것 (설계 문서에 먼저 반영, 커밋 `585d2be`)
- 로그인 유지 세션은 서버 기한만 늘리면 쿠키가 로그인 30일 뒤 사라진다 → 연장할 때 쿠키 `Max-Age`도 재발급.
- 401 한 곳 처리가 로그인 화면의 "비밀번호 틀림" 401까지 잡아 로그인 화면으로 튕긴다 → 로그인 · 가입 요청의 401은 제외.
- `?next=/\evil.com`(브라우저가 `//`로 읽음), `/\t/evil.com`(탭을 지우고 읽음)이 뚫린다 → `/\` · 제어 문자 거절 + 해석한 출처 확인.
- Windows에는 IANA 시간대 DB가 없어 `time.LoadLocation`이 실패 → `time/tzdata`를 실행 파일에.
- 비밀번호 변경의 "현재 비밀번호"가 무차별 대입 통로 → 사용자별 시도 제한(`rl:pwchange:user:`).
- 요청 제한 횟수가 문서에 없었다 → 표로 정함. 만료 세션 정리는 M7까지 쌓임 → 로그인 때 그 사용자 것 정리.
- 기본 잠금을 openapi `security`로 강제(새 경로는 저절로 잠김) + `TestDefaultLock`. 비밀번호 12 ~ 128자, 본문 1MiB, sqlc는 Docker로.
- 하지 않은 것: 세션 절대 수명 · 폼 라이브러리(superforms, M5에서 다시 봄) · `__Host-` 쿠키 접두사(M0 결정 유지) · 가입 409의 가입 여부 노출(M8에서).

### [설계] 화면 문구를 반말 → 부드러운 존댓말(해요체)로 (사용자 결정)
- DESIGN 8장 규칙을 먼저 바꾸고, DESIGN 4 · 7장 문구 예시 · REQUIREMENTS R1 알림 문구 · 목업 문구 · API 오류 message · 프론트 화면 문구를 같이 바꿨다. "~해 주세요" · "~했어요"처럼, 합쇼체와 반말은 쓰지 않는다. 인사는 "OO님, 안녕하세요".
- 목업 안의 "자동 사고" 예시 문장(사용자가 쓴 글이라는 설정)은 사람의 혼잣말이라 그대로 뒀다. CLI 로그 · 관리 명령 오류는 개발자용이라 평서문 그대로.

### [설계] 구현 중에 바꾼 것
- 바깥 라우팅을 mux가 아니라 경로 앞부분으로 나눈다: 메서드 없는 `/api/`와 `GET /`를 한 mux에 두면 Go가 패턴 충돌로 거부한다(ARCHITECTURE 1절).
- 413(`PAYLOAD_TOO_LARGE`)과 오류 `code` 목록을 ARCHITECTURE 3.2절에.
- 관리 명령을 `task user -- …`로도 부를 수 있게(Taskfile, `interactive: true`).

### 막혔던 일
**openapi에 `security`를 적었는데 생성 코드가 아무것도 안 잠금**
- 증상: 생성 코드에 `CookieAuthScopes`가 없었다 — 인증 미들웨어가 볼 표시가 없음.
- 원인: oapi-codegen std-http 템플릿은 `compatibility.enable-auth-scopes-on-context`가 켜져 있을 때만 표시를 넣는다(템플릿 소스에서 확인).
- 해결: `api/oapi-codegen.yaml`에 옵션 추가 → 잠긴 5개 경로에만 표시. CLAUDE.md "함정"에 "끄지 말 것".

**생성 코드가 `github.com/oapi-codegen/runtime`을 요구**
- 원인: 요청 본문이 있는 경로가 처음 생겨 `runtime.BindJSONBody`를 쓴다(M1 DEVLOG에 예고한 그대로). 사용자가 설치했고, 이후 같은 경우는 Claude가 설치하고 보고하기로 했다(CLAUDE.md "작업 방식").

**`panic: pattern "GET /" conflicts with pattern "/api/"`**
- 원인: Go 1.22 mux는 "메서드는 더 좁은데 경로는 더 넓은" 두 패턴을 우열을 못 가린다며 등록 단계에서 거부한다.
- 해결: 맨 바깥은 `strings.HasPrefix(path, "/api/")`로 나누고, 각각 따로 mux.

**SvelteKit 3 `resolve('/')`가 타입 오류**
- 원인: `resolve()`는 라우트 id를 받는데, 라우트 묶음 `(app)`이 id에 들어간다.
- 해결: `resolve('/(app)')` · `resolve('/(app)/settings')`. 같이 `goto`의 `replaceState`가 지원 중단 경고 → `replace`. CLAUDE.md "함정"에.

**vitest `test` 설정이 `vite`의 `defineConfig` 타입에 없음**
- 해결: `vitest/config`의 `defineConfig`를 쓴다(vite 설정에 test 칸을 더한 것).

**참고: 피어 의존성 경고 2개**
- `openapi-typescript` → TypeScript `^5`(설치는 6), `runed`(bits-ui 의존) → `@sveltejs/kit ^2`(설치는 3). `pnpm check` · build · 화면 확인이 모두 정상이라 그대로 둔다. 업데이트가 나오면 다시 본다.

### 이번 세션 커밋
`f236f0b` CLAUDE.md 사전 검증 목표 / `585d2be` 설계 문서 / `5c9d25a` sqlc · testutil / `c0d22cb` 마이그레이션 / `9ecda22` 인증 핵심 / `029e124` 설정 · IP / `5f5edae` 인증 API / `bd14677` 관리 명령 / `a628517` 프론트 / `788d546` task test / `d66b5fd` CI / (마지막 커밋) 유출 확인 테스트 보강 · `task user` · 문서

### 다음에 할 일
- 집 PC: 위 "현재 상태"의 "기기별로 할 일" · "집 PC에서 할 확인" → PLAN에서 M1 · M2 ✅
- M3 첫 배포: 배포 환경 결정(결정 기록 0007) — 개념 설명(서버 · 도메인 · HTTPS · 리버스 프록시)부터, 후보를 비교해 사용자가 고른다
