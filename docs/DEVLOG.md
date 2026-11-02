# 개발 일지 (v2)

> 겪은 일 · 막힌 일 · 그때 배운 것을 적는다. **최신이 위.** 지난 기록은 고치지 않고 위에 덧붙인다.
> 로드맵 · 진행 상황은 [PLAN.md](PLAN.md), 결정의 이유는 [decisions/](decisions/README.md). v1 일지는 [archive/DEVLOG_v1.md](archive/DEVLOG_v1.md).

## 현재 상태
| 항목 | 내용 |
|---|---|
| 단계 | **M5 운영 배포 · PC 확인 끝(2026-11-02), 휴대폰 → PC 두 기기 확인 남음** → 그다음 **M6 기록 목록 · 통계** (PLAN 2장) |
| 끝난 것 | **M5 사고 기록**(2026-11-02, 운영 배포 · PC 화면 확인, CI 통과 — 휴대폰은 아직): 왜곡 seed 11 · 초안 API(Valkey, rev · 10개 · Lua) · 기록 API · S-04 8단계 · S-06 · S-02 이어 쓰기, 시각 고르기(`DateTimePicker`) · 머리 `←`. **PC 배치**(2026-11-01): 사이드바 · 화면별 2단(64rem 이상) — DESIGN 3.2절. **M4 체크인**(2026-11-01, 운영 배포 · 휴대폰 5 ~ 6초): 감정 seed 10 · 체크인 API 6개(그날 = 사용자 시간대), S-02 오늘(체크인 원 · 오늘 목록) · S-03 작성 · 수정 · 삭제 확인, `apierr` 공용 패키지. **스타일 D2**(미색 종이 · 잉크 남색 · 제목만 고운바탕) · 로고(비켜 선 두 원) · 부제 "한 걸음 떨어져, 내 마음을 다시 보기" — DESIGN 6장. M3 첫 배포: `https://cbt.sagming.com`, 완료 기준 7개(휴대폰 · PC 로그인, DB · Valkey 포트 막힘, 실제 IP 요청 제한, 복원, README대로 재배포, 재부팅 복구, 실패 시 자동 되돌림) — SERVER 4.7 ~ 4.19절. M0 · M1 · M2 전부(회사 · 집 PC에서 확인). M2 인증: `users` · `sessions`, 세션 쿠키 로그인(로그인 유지 · 쿠키 재발급 · Valkey 캐시), 요청 제한 4종, 가입(플래그) · 내 정보 · 비밀번호 변경 · 다른 기기 로그아웃, 유출 비밀번호 거절, 기본 잠금(openapi security), `task user` 관리 명령, 프론트 로그인 · 가입 · 설정 · 401 처리 · 메뉴, Vitest, CI 통합 테스트 · 생성 코드 확인 |
| 진행 중 | 없음 |
| 다음에 할 일 | 운영은 시각 고르기 · 뒤로 버튼까지 올라가 있다(회사 PC `727b1eb`, 그 뒤 커밋은 문서뿐). **1. 사용자**: 휴대폰 1 ~ 3단계 → [나중에 이어 쓰기] → PC에서 4단계부터 완료 — 키보드 가림 · 달력 · 시 · 분 선택도 같이 본다. 되면 PLAN M5 완료 기준을 체크한다. **2. M6 기록 목록 · 통계**: 구현 전에 사전 검증 → 설계 · 순서 · 의존성을 먼저 보여 준다(최신 블록 "M6 시작할 때"). **사용자 숙제**(M6와 나란히 해도 됨): 인지 왜곡 11개 설명 · 예시 → 새 마이그레이션으로 넣는다, 단계 제목 · 도움말 → `web/src/lib/record-form.ts`의 `STEPS` |
| **서버에 접속되는 기기** | **집 PC · 노트북 · 회사 PC**(노트북 2026-10-31, 회사 PC 2026-11-02 추가 — 열쇠 메모 `home-pc` · `laptop` · `company-pc`). 셋 다 `ssh cbt` · `task deploy` · `task backup:pull` 가능. 새 기기는 SERVER.md 3.4절 |
| 기기별로 할 일 | **집 PC · 노트북**: `git pull` 뒤 `task setup`(lucide 버전 올라감) · `task migrate`(00004 사고 기록) — 2026-11-02. **집 PC**: 2026-11-01분 `@fontsource/gowun-batang` · 00003도 위 두 명령에 같이 들어 있다. 집 PC 전역 설정(`~/.claude/CLAUDE.md`)의 공용 인프라 절을 드라이브 사본 `laptop_CLAUDE.md`에 맞추기(첫 배포 · `laptop` 열쇠 · 경로 표기). 백업 개인키 `$HOME\cbt-backup.key`를 노트북에서 USB로 옮기기(SERVER 4.13절). **모든 PC**: `~/.ssh/config`에 별명 `cbt`(SERVER.md 3.2절, `task deploy` · `task backup:pull`이 씀) · `age` 설치 — README "운영 서버를 다루는 PC라면". 셸 · systemd 파일을 직접 칠 기기는 에디터 상태바를 "Indent Using Tabs"로(공백이 섞였던 일 — 2026-10-31 (2)). **회사 PC**: 서버 준비 끝남(2026-11-02 — 열쇠 `company-pc` · 별명 `cbt`(도메인) · `age` 1.3.2). 백업 개인키는 필요해지면 USB로. **노트북**: 다 끝남(2026-10-31 — 처음 한 번 · 열쇠 `laptop` · 별명 `cbt` · `age` 1.3.2) |
| **기억할 날짜** | **2026-11-30 Vultr 크레딧 $250 만료** — 그 전에 테스트 서버 삭제(Destroy, 정지는 요금이 계속 나감), 12월부터 서버 한 대 값이 카드로. **같은 날 Vultr 자동 백업($1/월) 유지 여부 판단**(우리 DB 백업이 믿을 만하면 끔). 2027-10-30 `sagming.com` 자동 갱신 (Cloudflare). 2029-04 Ubuntu 24.04 지원 종료 → 새 서버로 옮기기 |

## 작성 규칙
- 세션(작업 단위)이 끝날 때 날짜 블록을 맨 위에 추가하고, 위 "현재 상태"를 고친다.
- 막혔던 일은 해결됐어도 남긴다. 형식: **증상 / 원인 / 해결 / 배운 것**.
- 구현 중에 설계를 바꿨으면 `[설계]` 항목으로 이유를 적고, 해당 설계 문서의 변경 이력도 고친다.
- 블록 끝에 **이번 세션 커밋**과 **다음에 할 일**을 꼭 쓴다.

---

## 2026-11-02 (3) · 시각 고르기 · 뒤로 버튼 (회사 PC, 자동 작업)

### 한 일
- 사용자가 M5를 운영에 배포해 PC 화면 정상 작동을 확인(휴대폰은 회사라 못 봄). 써 보고 준 의견 2개를 반영했다.
- **시각 입력 → `DateTimePicker`**: 브라우저 기본 `datetime-local`은 달력 아이콘만 눌러야 열린다(사용자 의견). 칸 전체가 버튼인 shadcn Popover + Calendar로 바꾸고, 시 · 분은 어디를 눌러도 열리는 `select`로. S-03 · S-04 1단계에 같이 적용.
- **머리 `×` → `←`**: 작성은 2단계부터 이전 단계 · 1단계에서 오늘, 수정은 어느 단계에서든 상세로, 상세 · 체크인은 오늘로.
- 고친 것: [나중에 이어 쓰기]가 8단계에서 빠져 있었다(DESIGN 4.3절은 "3단계 이후 모든 단계"). `←`가 한 단계씩 가므로 8단계에서 나갈 길이 더 필요했다.
- 설치: `shadcn-svelte add calendar popover`(`pnpm dlx`로 CLI를 임시로 받음). 이때 `@lucide/svelte`가 ^1.63 → ^1.66으로 같이 올랐다. `button`은 내용이 같아 안 바뀌었다.

### 사전 검증에서 바꾼 것
- [설계] `←`를 `history.back()`으로 하지 않고 갈 곳을 정해 뒀다: 완료 뒤 기록에 지워진 초안의 단계 주소가 남아 있어 그쪽으로 돌아가면 "찾을 수 없어요"가 뜨고, 첫 화면에서 누르면 앱 밖으로 나간다.
- [설계] 달력에서 오늘 뒤의 날짜는 못 고른다(DATA_MODEL 3.4절). 오늘 안의 미래 시각은 전처럼 저장할 때 서버가 알려 준다.
- [설계] 달력이 열린 동안 체크인의 `Esc`는 달력만 닫는다(페이지를 나가지 않게).

### 막힌 것
- **증상**: 달력 제목이 "11월 2026년". **원인**: 생성 부품의 제목이 "월 연" 순서로 고정이다. **해결**: 생성 코드를 고치지 않고 `monthFormat`에 연 · 월을 다 넣고 `yearFormat`을 비웠다. **배운 것**: 화면 낭독기용 달력 이름의 기본값도 영어 "Event"라 `calendarLabel="날짜"`를 줬다.
- **증상**: `shadcn-svelte add`가 "덮어쓸까요"에서 멈춤. **원인**: 달력이 `button`에 기대는데 이미 있어서 묻는다. **해결**: `--yes --overwrite`로 받은 뒤 `git status`로 `button`이 안 바뀐 것을 확인.

### 확인한 값
- `task lint` · `task test` · `task build` 통과, Vitest 41개(`formatLocalInput` 1개 추가).
- 브라우저(로컬, 지어낸 글): 칸의 글자 부분 클릭 → 달력 열림(1280px), 11월 1일 클릭 → "11월 1일 (일)", 미래인 14일 클릭 → 그대로, 09시로 바꾸면 초안 `occurred_at` `2026-11-02T00:22:00Z`(KST 09:22), 글을 안 쓰면 초안이 안 생김(`?step=1` 그대로). 제목 "2026년 11월" → 이전 달 "2026년 10월" → 닫았다 열면 다시 11월. 375px: 달력이 화면 안(20 ~ 257px), 가로 스크롤 없음.
- `←`: 2단계 → 1단계 → `/`(저장돼 있으니 확인 창 없음), 수정 `?step=5` → 상세 → `/`.
- **확인 못 한 것**: 고친 게 없는 체크인 수정 화면에서 달력을 연 채 `Esc`(코드만 봄), 실제 휴대폰(기기 선택창), 운영 배포.

### 이번 세션 커밋
- `docs(M5): 시각 고르기 · 뒤로 버튼 설계` · `feat(M5): 칸 전체를 누르는 시각 고르기` · `feat(M5): 머리에 뒤로 버튼` · `docs(M5): 시각 고르기 · 뒤로 버튼 기록`

### 운영 배포 뒤 확인 (Claude가 바깥에서 본 값)
- 사용자가 `task deploy`. `https://cbt.sagming.com/` 200, `/api/health` postgres · valkey `ok`, 비로그인 `/api/drafts` 401. 이 응답으로는 버전을 알 수 없다(버전은 `ssh -t cbt "sudo cbtctl version"`).
- GitHub Actions: `727b1eb` success(코드 커밋이 같이 push된 것). 회사 PC에는 `gh`가 없어 공개 API로 봤다.

### 다음에 할 일
- (배포함: 회사 PC `727b1eb`) **사용자**: 휴대폰 1 ~ 3단계 → PC 이어 쓰기(PLAN M5 완료 기준). 왜곡 문구 · 단계 제목 다듬기.
- 그다음 **M6 기록 목록 · 통계**.

### M6 시작할 때 (이번에 코드를 보며 알아 둔 것 — 설계는 아직 안 함)
- **`S-06`의 `←`는 지금 오늘 화면으로 간다**: 목록(`/records`)이 생기면 어디로 돌아갈지 다시 정한다(DESIGN 4.5절). 수정 · 작성의 `←`는 그대로.
- **메뉴**: `(app)/+layout.svelte`의 `nav`에 기록 · 통계를 더한다. 쓰는 중 메뉴를 숨기는 `focusMode`는 `/edit`로 끝나는 주소 · `/records/new` · `/checkin` 기준이라, 새 주소가 이 규칙에 걸리지 않는지 본다.
- **있는 것**: 그날 목록 `GET /api/records?date=` · `GET /api/checkins?date=`, 그날 범위 `internal/localday`, 통계 버전 `cache.BumpStatsVersion`(저장 · 수정 · 삭제 때 이미 올린다). S-02의 "변화가 가장 큰 감정 + 외 N"은 `(app)/+page.svelte`의 `biggest()`에 있으니 S-05 카드와 같이 쓰게 뽑는다.
- **PLAN의 "정기 백업 켜기"는 M3에서 이미 켰다**(SERVER.md `enable --now cbt-backup.timer`). M6에서는 파일이 하루 한 번 생기는지만 확인하면 된다.
- **새 의존성**: 차트(DESIGN 5.1절 Chart, decisions 0003 LayerChart) · E2E Playwright. 큰 작업 규칙대로 시작할 때 목록을 먼저 보여 준다.

### 알아 둘 것
- 크롬 뒤로 가기로 완료 전 단계 주소에 돌아가면 "이 초안을 찾을 수 없어요"가 뜬다(화면의 `←`는 이 경로를 타지 않는다). 내용 손실은 없어 그대로 뒀다.
- Valkey가 죽어 있으면 저장 실패 표시 · 나가기 확인까지 약 7초 걸린다(go-redis 재시도). 드문 경우라 그대로 뒀다.
- 회사 PC 로컬 DB에는 확인용 계정 `m5-test-…@example.com`과 지어낸 기록이 있다(ALLOW_SIGNUP을 그때만 켜서 만듦, 운영과 무관).

## 2026-11-02 (2) · M5 사고 기록 구현 (회사 PC, 자동 작업)

### 한 일
- 마이그레이션 00004: `distortions`(seed 11개) · `thought_records` · `record_emotions` · `record_distortions`.
- API(`internal/record`, openapi 0.4.0): 초안 만들기 · 저장 · 목록 · 불러오기 · 지우기, 완료 · 상세 · 수정 · 삭제 · 그날 목록, 왜곡 목록.
- 공용으로 뺌: "그날" 계산 `internal/localday`, 통계 버전 올리기 `cache.BumpStatsVersion`(체크인 동작 변경 없음).
- 화면: S-04 8단계 · 자동 저장(`draft-saver.ts`) · 앞에서 쓴 답 · D-02 · [나중에 이어 쓰기], 수정 모드, S-06, S-02(사고 기록 쓰기 · 이어 쓰기 · 오늘 목록에 기록).
- 폼 라이브러리를 쓰지 않기로 정식 결정(decisions 0008). 옛 결 목업 `Record.*.html`은 다시 그리지 않음.

### 사전 검증에서 바꾼 것
- [설계] **같은 초안 두 번 완료 → 기록 하나**: `UNIQUE (user_id, source_draft_id)` + `ON CONFLICT DO NOTHING`. 처음엔 먼저 찾아보려 했으나 동시 요청이 둘 다 "없다"고 본다. 사용자별 UNIQUE라 남의 초안 id와 부딪히지 않는다.
- [설계] **개수 검사 · rev 비교는 Lua로 저장과 한 덩어리**. 새 rev를 넣은 JSON을 서버가 만들어 넘긴다 — Lua cjson으로 다시 만들면 빈 배열 `[]`이 `{}`로 바뀐다.
- [설계] **완료는 화면 내용으로**(Valkey를 읽지 않음). 초안이 날아가도 화면에 보이는 것이 저장된다. 대신 완료 전에 남은 저장을 먼저 마쳐, 다른 기기가 먼저 고쳤으면 D-02로 멈춘다.
- [설계] 8단계 "후"는 "전" 값에서 시작, 8단계 화면이 곧 요약(따로 요약 화면 없음), 일어난 시각은 화면을 연 때로 정해 초안에 넣음(이틀 뒤 이어 써도 안 밀리게), 수정 모드는 어느 단계에서든 [저장] + S-06 칸마다 [고치기].
- [설계] 초안이 없어졌을 때(다른 기기에서 완료 · 14일 만료) 쓴 내용을 버리지 않고 [새 초안으로 저장] / [오늘로 돌아가기]. Valkey 장애는 초안 API만 503.

### 막힌 것
- **증상**: 주소가 `?step=5`인 빈 화면에서 첫 글자를 치자 2단계로 튀었다(브라우저 확인에서 발견). **원인**: 단계를 `$derived(clampStep(주소, form))`으로 둬서, 글을 쓰면 "들어갈 수 있는 곳"이 넓어져 같이 바뀌었다. **해결**: 단계는 상태로 두고 이동이 끝난 뒤(`afterNavigate`)에만 주소를 다시 읽는다. 처음 열 때 주소도 보정한 값으로 바꾼다. **배운 것**: 파생값은 "입력이 바뀌면 같이 바뀌어도 되는가"를 먼저 묻는다.
- **증상**: 이어 쓰기 초안 [지우기]를 눌러도 안 지워졌다(브라우저 확인에서 발견). **원인**: `AlertDialog.Action`이 창을 먼저 닫아 `onOpenChange`가 대상을 비운 뒤 `onclick`이 돌았다. **해결**: 대상을 상태로 들고 있는 창에서는 `Button`을 쓴다. CLAUDE.md "함정"에 적음.
- **증상**: 타입 오류 `keepFocus` · `noScroll`이 없다. **원인**: SvelteKit 3은 `goto` 옵션이 `reset: false` 하나로 합쳐졌다. **해결**: 초안이 생겨 주소를 바꿀 때 `{ replace: true, reset: false }` — 쓰는 중에 커서 · 스크롤이 안 움직이게. CLAUDE.md "함정"에 적음.
- **증상**: 긴 명령이 `ENAMETOOLONG`으로 실패. **원인**: Windows 명령줄 길이 제한. **해결**: 파일은 Write로 나눠 쓴다(작업 방식 문제, 코드와 무관).

### 확인한 값
- `go test ./...` 통과. M5 통합 테스트 16개 PASS(건너뜀 없음을 `-v`로 확인): 오래된 rev 409 · 동시 저장 8개 중 성공 1 · 11번째 409 · 동시 생성 15개 → 10개 · 유령 초안 빠짐 · TTL 14일 갱신 · 남의 기록 · 초안 404 · 완료 뒤 초안 삭제 · 두 번 완료 200 · 동시 완료 6개 → 기록 1개 · 입력 오류와 트랜잭션 도중 DB 실패(트리거로 재현) 둘 다 초안 남음 · 시간대 경계 · Valkey 죽음 → 초안만 503, 완료 201.
- Vitest 40개 통과(`draft-saver` 11 · `record-form` 11 추가). `task lint` · `task build` 통과.
- **Valkey 재시작**(PLAN 완료 기준): 초안 키 `draft:*` 1개 · `drafts:*` 1개, TTL 1209598초 → `docker compose restart valkey` 뒤에도 그대로, `appendonly yes`.
- 브라우저(로컬, 지어낸 글, 1280 · 375px): 첫 입력 → `?draft=…&step=1`로 바뀌어도 커서가 칸에 남음 · 저장됨 · 방금, [나중에 이어 쓰기] → S-02 "3/8 · 방금" → 다시 열면 3단계, 8단계까지 Ctrl+Enter, 후 강도 50 → 50에서 시작, 다른 기기 흉내(PUT) → 409 → D-02 → [그쪽 내용 불러오기]로 반대 근거가 들어옴, 완료 → S-06 · S-02 "불안 50 → 20 외 1", `?step=5` 빈 초안 → 1단계, 375px 가로 스크롤 없음.
- Valkey를 끄고 쓰기 → "저장 안 됨 · 다시 시도" + 안내, 닫기 → 확인 창(약 7초 뒤, go-redis 재시도 때문), 켜고 다시 시도 → 저장됨(rev 2).
- **확인 못 한 것**: 휴대폰 실기기(키보드 · `interactive-widget` · 빠른 조작), 실제 두 기기, 운영 배포, 한글 조합 중 Ctrl+Enter 무시는 코드만 봤다. E2E는 M6 완료 기준이다.

### 이번 세션 커밋
- `docs(M5): 사고 기록 사전 검증 반영` · `refactor(M5): "그날" 계산 · 통계 버전 올리기를 공용으로 뺌` · `feat(M5): 인지 왜곡 · 사고 기록 테이블과 쿼리` · `feat(M5): 사고 기록 · 초안 API` · `feat(M5): S-04 사고 기록 8단계 · S-06 상세 · S-02 이어 쓰기` · `docs(M5): 구현 기록`

### 다음에 할 일
- **사용자**: `task deploy` → 휴대폰으로 1 ~ 3단계 → [나중에 이어 쓰기] → PC에서 4단계부터 이어 써서 완료. 되면 PLAN M5 완료 기준 체크.
- **사용자**: 인지 왜곡 11개 설명 · 예시, 단계 제목 · 도움말(`web/src/lib/record-form.ts`의 `STEPS`) 다듬기. 왜곡 문구는 주면 새 마이그레이션으로 넣는다.
- 그다음 **M6 기록 목록 · 통계**.

## 2026-11-02 · 회사 PC 서버 준비 (회사 PC)

### 한 일
- 회사 PC에서도 배포 · 백업을 하려고 서버 접속을 준비(SERVER.md 3.6절). 명령은 사용자가 직접 쳤다.
- 회사 네트워크가 22번을 막는지 먼저 확인 → 열려 있음(`Test-NetConnection … -Port 22` → `True`). 핫스팟이 필요 없다.
- 열쇠 `company-pc`(ed25519, passphrase) → Vultr 웹 콘솔에서 `authorized_keys`에 한 줄 덧붙임 → `age` v1.3.2 → `~/.ssh/config`에 별명 `cbt`.
- [설계] 별명의 `HostName`을 **IP 대신 도메인**(`cbt.sagming.com`)으로 적었다: Cloudflare가 DNS만이라 바로 서버로 가고, IP를 어디에도 옮겨 적지 않아도 된다. SERVER.md 3.2절에 서버 지문(ED25519)을 적어, 새 기기에서 첫 접속 때 `yes` 전에 비교하게 했다(지문은 공개 정보).

### 확인한 값
- `ssh -G cbt` → `hostname cbt.sagming.com` · `user sagming40` · `identityfile ~/.ssh/cbt_vultr`.
- `ssh cbt` → 지문 `SHA256:hriG4Hqp…` 확인 뒤 `yes` → passphrase → `sagming40@cbt-1:~$`.
- **회사 PC 첫 배포**(SERVER 4.20절): `task deploy` → 배포 직전 백업 · 마이그레이션 없음 · 재시작 · 헬스체크 통과(2번째) → `sudo cbtctl version` `66007b5d4815` → `3dd4e15af789`, `curl -sI` `200`. 문서만 바뀐 상태라 앱 동작은 같다(배포 길 시험).
- 헛디딘 것: `ssh cbt "sudo cbtctl version"` → `sudo: a terminal is required …`. 한 줄 실행은 서버에 터미널이 안 붙어 sudo가 비밀번호를 못 묻는다 → `ssh -t`(SERVER 5장 머리에 적음, 안내한 명령이 틀렸던 것).

### 이번 세션 커밋
- `docs(M4): 회사 PC 서버 준비 기록` · `docs(M4): 회사 PC 첫 배포 기록`

### 다음에 할 일
- PC 배치는 어제 노트북에서 이미 운영에 올라가 있었다(`66007b5`). 다음은 **M5 사고 기록** — 사전 검증 → 설계 · 순서 · 의존성 먼저 보여 주기. 그 전에 정할 것: 사고 기록 목업(`Record.*.html`)을 D2로 다시 그릴지.

## 2026-11-01 (5) · PC 배치 반영 — 사이드바 · 2단 (노트북, 자동 작업)

### 한 일
- 사용자가 메뉴를 **사이드바**로 정함 → DESIGN 3.2절(틀 68rem · 화면별 2단 · 글줄 40rem 유지), 4.1 · 4.2 · 4.4 · 4.6 · 4.7 · 4.8 "데스크톱", 4.2 키보드, 8장 단축키 규칙을 먼저 고침.
- 사이드바: 로고 · 부제, 메뉴, 맨 아래 [+ 체크인] · 닉네임 · 시간대, 스크롤해도 제자리. 모바일 하단 탭은 따로 떼어 그대로.
- S-02 오늘: 왼쪽 종이(체크인 원) | 오른쪽 오늘 목록, 감정 이름 명조, 0건이면 흐린 로고 + 안내. "이번 주" 7칸 · 요약 타일은 날짜별 개수 API가 필요해 M6로.
- S-03 체크인(작성 · 고치기 같이): 왼쪽 칩 | 오른쪽 종이(강도 · 메모 · 시각 · 저장). `EmotionPicker`에 `part`(chips · sliders)를 더해 같은 값을 두 자리에 묶음. `Ctrl(⌘)+Enter` 저장 · `Esc` 닫기.
- S-01 로그인 · 가입: `AuthShell` 틀 — 왼쪽 반에 부제 · 세 줄 소개 · 흐린 두 원, 오른쪽 폼.
- S-08 설정: 목차(지금 보는 묶음에 선) | 묶음마다 종이 한 장.

### 사전 검증에서 바꾼 것
- [설계] **체크인 단축키 `C` 뺌**: 글자 하나 단축키는 화면 낭독기 키와 부딪히고 WCAG 2.1.4가 끄는 설정을 요구한다 → 조합키 · Esc만 (DESIGN 8장).
- [설계] **고친 게 있으면 Esc로 닫지 않음**: 한글 입력을 취소하려 Esc를 누르다 고른 감정 · 메모를 날릴 수 있다. 한글 조합 중(`isComposing`)에는 Ctrl+Enter도 무시(마지막 글자가 아직 확정 전).
- 삭제 확인 창이 열려 있을 때 Esc는 창만 닫는다(페이지는 그대로).

### 확인한 값
- `task lint` 0 issues · svelte-check 0 errors, Vitest 16개 통과(Go 코드는 안 바뀜).
- 내장 브라우저 1440 · 1024 · 375px: 1440에서 사이드바 240 + 틀 1088px, 체크인 칩 5열 | 종이. 빈 폼 Esc → 오늘로, 짜증을 고른 뒤 Esc → 그대로, Ctrl+Enter → 저장 · 토스트 "기록했어요" · 목록 1건. 고치기에서 확인 창 열고 Esc → 창만 닫힘. 가입 화면 반으로 나뉨. 375px는 전과 같음(하단 탭 · 아래 고정 저장). 콘솔 오류 0. 시험 체크인은 로컬 DB에서 지움.
- 확인 못 한 것: 운영 배포 뒤 실제 PC 모니터 · 휴대폰, 라이트 테마 PC 화면(개발 브라우저가 다크).

### 막혔던 일
- **1024px에서 감정 칩 글자가 세로로 꺾임** — 원인: 2단을 `md:`(48rem)부터 걸어서, 사이드바를 빼면 27rem 남짓인 본문을 다시 둘로 나눔. 해결: 2단은 `lg:`(64rem)부터, 칩 5열 격자는 `xl:`(80rem)부터. 배운 것: 사이드바가 생기는 너비와 2단이 되는 너비는 따로 정해야 한다(DESIGN 3.2절에 적음).

### 이번 세션 커밋
- `docs(M4): PC 배치 확정 — 사이드바 · 화면별 2단` · `feat(M4): PC 배치 — 사이드바 · 2단 화면 · 체크인 단축키` · `docs(M4): PC 배치 구현 기록`

### 다음에 할 일
- `task deploy`로 운영에 올리고 PC 모니터 · 휴대폰으로 확인 → M5 사고 기록.

## 2026-11-01 (4) · PC 화면 D2 목업 (노트북)

### 한 일
- PC 화면이 3.2절 "본문 최대 40rem 한 줄"이라 넓은 모니터에서 가운데 띠 하나만 남아 부실해 보였다(사용자 의견). PC 목업은 옛 결의 `Record.desktop.html` 하나뿐이었다.
- 참고: cosmic-watch-explorer 실행 화면(반으로 나눈 로그인 · 큰 제목 + 타일 · 목록 옆 패널 · 필터 막대 · 카드형 설정), PlayLedger · 자재 사전 목업.
- `docs/mockups/Desktop.html` 보드 하나에 8화면: S-01 로그인(반으로 나눔) · S-02 오늘(원 · 이어 쓰기 | 오늘 일기, 이번 주 7칸) · 0건 · S-03 체크인(고르기 | 세기 · 메모 · 저장) · S-04 5단계(본문 | 앞에서 쓴 답) · S-05 목록 + S-06 상세(목록 옆에 바로 펼침) · S-07 통계 · S-08 설정(목차 | 카드). 메뉴는 사이드바 / 위 메뉴 두 안, 다크.
- PC 단축키 제안(체크인 `C`, 저장 `Ctrl+Enter`, 닫기 `Esc`)을 화면 설명에 적었다. 확정 전이라 DESIGN 본문은 아직 안 고쳤다(9.1 계획 표 · README만).

### 확인한 값
- 내장 브라우저로 8화면 · 사이드바 / 위 메뉴 · 라이트 / 다크를 열어 봄. 콘솔 오류 0. 체크인 칩을 누르면 슬라이더가 50에서 생기고, 0개면 저장이 꺼진다.

### 막혔던 일
- **cosmic 실행 실패** — 증상: `did not find executable at 'C:\ProgramData\anaconda3\python.exe'`, pip는 `Fatal error in launcher`. 원인: `backend\venv`가 다른 PC(anaconda, `C:\__dev_repos\…`)에서 만든 걸 폴더째 복사한 것. 해결: venv를 지우고 노트북의 Python 3.13으로 다시 만듦. 배운 것: venv는 만든 PC · 경로에 묶여서 옮기면 못 쓴다 — 기기마다 새로 만든다.

### 이번 세션 커밋
- `docs(M4): PC 화면 D2 목업`

### 다음에 할 일
- 같은 날 (5)에서 이어 감.

## 2026-11-01 (3) · M4 체크인 API · 화면 (노트북, 자동 작업)

### 한 일
- 사전 검증 → 설계 · 순서 · 의존성을 먼저 보여 주고 사용자 확인 뒤 자동으로(CLAUDE.md "구현 전 사전 검증").
- 문서 먼저(`12612c3`): ARCHITECTURE 2절(`apierr`, 사용자 id를 함수로 넘김), DATA_MODEL 3.4절(한 트랜잭션 · 나중 저장이 이김 · 그날 범위 · 미래 5분).
- 리팩터링(`d6584a3`): 오류 봉투 · `ValidationError`를 `internal/apierr`로 — `checkin`이 `auth`를 import하지 않게. `auth.ValidationError`는 별칭으로 남김. 동작 그대로(기존 테스트 전부 ok).
- API(`85052a9`): 마이그레이션 00003(감정 seed 10 · 체크인 · 체크인 감정 · 인덱스), openapi 0.3.0 경로 6개, `internal/checkin` service · handler, server는 `checkin.Handler`를 별칭으로 임베드(두 기능 모두 `Handler`라 필드 이름이 겹쳐서). 저장 · 고치기 · 지우기 때 `stats:ver` +1(TTL 30일, Valkey 실패는 로그만).
- 화면(`7106180`): S-02 오늘(체크인 원 · 오늘 목록 · 0건 안내, 서버 상태 칸은 뺌), S-03 작성 · 수정(감정 칩 → 강도 50, 기본 range 슬라이더, 밑줄 메모, 시각 바꾸기, 아래 고정 [저장], 하단 탭 숨김), D-01 삭제 확인, `lib/tz.ts`(사용자 시간대 벽시계 ↔ UTC).
- 사용자가 `pnpm dlx shadcn-svelte@latest add alert-dialog --cwd web`(기존 `button`은 덮어쓰지 않음 — 이번 작업과 상관없는 변경이 섞이지 않게). 새 npm 패키지 없음.
- [설계] DESIGN 3.2절: 체크인(S-03) 중에도 하단 탭 숨김(고정 [저장] 자리), 데스크톱 사이드바는 그대로.

### 확인한 값
- `go test ./internal/server` 24 PASS · 0 SKIP — 새 테스트 6개: 감정 seed 순서, CRUD(감정은 화면 순서 · 메모 공백 다듬기 · 지우면 감정 행도 0), 입력 오류 8가지 400 + 경계값(강도 0 · 100, 메모 200자, 4분 뒤) 201 + 400이 반쪽 행을 안 남김, 남의 것 GET · PUT · DELETE 404 · 목록에 안 섞임 · 틀린 id 모양 404, 시간대(서울 ↔ 뉴욕으로 바꾸면 날짜 이동, 뉴욕 2026-11-01 서머타임이 끝나는 25시간 날 경계 03:59Z / 04:00Z / 다음 날 04:59Z / 05:00Z, 시간순), 통계 버전 3 · TTL 있음. `TestDefaultLock`에 새 경로 6개.
- Vitest 16(tz 7: 서머타임 시작 · 끝, 두 번 있는 01:30은 앞의 것). `task lint` 0 issues, svelte-check 0 · 0.
- 로컬 브라우저(390px, 라이트): 0건 안내 → 원 → 체크인(탭 숨김, 감정 고르기 전 저장 비활성) → 불안 · 짜증, 키보드 Home 0 · PageUp ×6 60 · → 65 → 저장 "기록했어요" → 오늘 `03:15 불안 65 ●●●○○ / 짜증 50 ●●●○○ / 메모` → 수정 화면에 값 그대로 · 시각 서울 03:15 → 짜증 빼고 "고쳤어요" → 지우기 확인 창 → "지웠어요" → 0건 → 지운 주소 "찾을 수 없는 기록이에요". (지어낸 테스트 체크인, 로컬 DB에서 지움)
- **운영 배포 · 휴대폰 10초 (M4 마지막 완료 기준)**: 사용자가 `task deploy`(`7d304a9`, 마이그레이션 00003이 운영에 처음 적용) → 휴대폰으로 앱 열기 ~ "기록했어요"까지 **5 ~ 6초**, 걸린 점 없음. Claude가 바깥에서: 로그인 없이 `/api/checkins` · `/api/emotions` → `401`(경로가 있고 잠김), health `200`. **M4 ✅.**
- **확인 못 한 것**: 다크 모드 체크인 화면은 브라우저로 안 봤다(토큰만 씀).

### 막혔던 일
**`testutil.Reset`이 seed 감정을 지움**
- 증상: 새 테스트에서 감정 목록이 비고 "없는 감정이 들어 있어요".
- 원인: `TRUNCATE users CASCADE`는 `users`를 참조하는 표를 **통째로** 비운다 — `emotions.user_id`(사용자 정의 감정 칸)가 `users`를 참조해서 기본 감정(user_id NULL)까지 사라졌다.
- 해결: `DELETE FROM users`(그 사용자 행에 딸린 것만 CASCADE).
- 배운 것: 시스템 데이터가 사는 표가 사용자 표를 참조하면 TRUNCATE CASCADE를 쓰면 안 된다.

**sqlc가 `unnest(a, b)`를 모름** — `SELECT unnest(a), unnest(b)`(같은 길이 배열은 짝지어 펼침)로.

**설계안의 틀린 말 두 개(Claude)**: "`TestDefaultLock`이 새 경로를 자동으로 확인" — 실제로는 손으로 적는 목록이라 직접 추가. "15:30Z = 뉴욕 11:30" — 그날 06:00Z에 서머타임이 끝나 10:30, 프론트 테스트가 잡음(`89822f0`).

**개발 서버 두 가지**
- air가 `gofmt -w` 뒤 파일 감시를 놓쳐 옛 실행 파일로 돌았다(새 경로가 404) → `task dev` 다시 띄움.
- 개발 서버가 켜진 뒤 생긴 alert-dialog 부품의 Tailwind 클래스가 안 만들어져 확인 창이 화면 밖(y=844)에 떴다 → 다시 띄우니 정상. 빌드는 매번 처음부터라 운영과는 상관없다.
- 배운 것: 새 부품을 깔았거나 생성 코드를 바꿨으면 개발 서버를 다시 띄우고 본다.

### 이번 세션 커밋
`12612c3` 설계 문서 / `d6584a3` apierr / `85052a9` 체크인 API / `89822f0` 테스트 주석 / `7106180` 체크인 화면 / `7d304a9` 문서 / (이 커밋) M4 끝

### 다음에 할 일
- M5 사고 기록(초안 자동저장 · 단계형 화면) — 사전 검증부터. 인지 왜곡 11개 설명 · 예시는 사용자 할 일(PLAN M5).

---

## 2026-11-01 (2) · 스타일 D2 · 로고 확정 (노트북)

### 한 일
- 사용자: 지금 화면(shadcn 기본)이 너무 밋밋하다 → 다른 프로젝트 목업(PlayLedger · 자재 사전 — 어두운 바탕 · 큰 굵은 제목 · 벤토 · 알약 메뉴)을 참고해 목업부터.
- `docs/mockups/Style.html`: 스타일 시안 A 밤의 일기장 · B 종이 노트 · C 고요한 물결 → 사용자 조합 D(B 바탕 + C의 가운데 원 체크인, 전부 명조) · D2(본문은 Pretendard) → **D2 채택**. S-02 · S-03 모바일, 라이트 / 다크.
- DESIGN 6장을 D2로: 미색 종이 바탕, 라이트 남색 `#4457C2` → `#33449E`(종이 위에서 떠 보여서), 서체 Pretendard + 고운바탕(제목만), 납작한 모양(선 · 점선 · 6px), 체크인 원(4초 숨 · 동작 줄이기면 멈춤), 목록 강도 점 5개. 4.1 와이어프레임 · 5.2 부품.
- 대비를 node로 다시 계산(WCAG 식) — 라이트 버튼 글자 8.53 · `muted` 5.12 · `input` 3.48, 다크 6.20 · 7.34 · 3.41. **다크의 강도 점(`accent`)이 바탕 대비 2.92**라 그림 기준 3:1 미달 → 다크 점은 `accent-text`(8.5)로.
- 부제: 사용자가 앱을 설명한 말("감정을 곰곰이 분석해 스스로 치유")에서 "치유 · 해소 · 곱씹다"는 피함(치료 대체로 읽힘 · 감정을 없애는 게 목표처럼 · 심리학의 반추) → 후보 중 **"한 걸음 떨어져, 내 마음을 다시 보기"** (사용자 선택). README 맨 위에.
- 로고: `Logo.html` 1차 1 ~ 4 → 부제를 그림으로 2차 5 ~ 9 → `LogoTrend.html` 상징 8 × 요즘 스타일 4(리퀴드 글래스 · 그레인 · 소프트 3D · 리소) → **5 · 비켜 선 두 원, 효과 없이 납작하게** (사용자 결정). 모양 수치 · 색 · 쓰는 곳을 DESIGN 6장 "로고 · 부제"에.
- 목업 보기: 노트북엔 Python이 없어 Node 내장 모듈로 만든 임시 정적 서버(스크래치 폴더, 저장소 밖)로 띄워 400 · 600 · 860px 폭에서 화면 확인.

### 막혔던 일
**브라우저 창이 `file://` 목업은 열기만 하고 화면을 못 찍음**
- 해결: 로컬 정적 서버(`http://127.0.0.1:5199`)로 열어서 확인. README의 `npx http-server`와 같은 일을 패키지 없이 했다.

### 앱에 D2 반영
- 사용자가 `pnpm -C web add -D @fontsource/gowun-batang`(5.3.0). Claude가: `layout.css` 토큰을 DESIGN 6장 D2 값으로(+ `--rule` · `--dot`, 다크 점은 `accent-text`), `--font-serif`, 그림자 변수 `--shadow-xs` 등을 0으로(생성 부품이 붙이는 그림자까지 전역으로 끔), `--radius` 7.5px → md 6px, 한국어 `word-break: keep-all`.
- `Logo.svelte`(비켜 선 두 원 + 선택 "CBT 저널" · 부제), 파비콘을 Svelte 로고 → 우리 상징(다크 브라우저면 `#A3AEF0`).
- 화면: 로그인 · 가입은 카드 없이 종이 위 로고 + 부제 + 명조 제목, 메뉴는 사이드바 로고 · 하단 탭 밑줄, 오늘은 사용자 시간대 날짜 + "오늘은 어떤 하루였나요?"(명조) + 서버 상태를 선 · 점선 구역으로, 설정은 카드 → 선으로 나눈 구역 · 명조 구역 제목, 보조 버튼 테두리 `input`.
- **체크인 원은 아직 안 넣었다** — 누르면 갈 `/checkin`이 없어서(없는 화면으로 가는 버튼이 됨). M4 체크인 때 `CheckinOrb`로.
- [설계] DESIGN 6장: 밑줄 입력은 기록 칸에만, 계정 폼은 상자 유지(입력칸인 게 바로 보이게) · 낱말 단위 줄바꿈.
- 확인한 값: `svelte-check` 0 · 0, `task lint` 0 issues · Prettier 통과, `task test` 종료 코드 0(Vitest 9). 브라우저(로컬 `task dev`, 사용자가 로그인): 로그인 · 오늘 · 설정을 모바일(390) 라이트 / 다크 · 데스크톱(1280) — 제목 `font-family` "Gowun Batang", 로드된 고운바탕 조각 10개. `pnpm build` 결과에 고운바탕 파일 375개(7MB, woff2 + woff) — 사용자는 화면 글자 조각만 받지만 **실행 파일은 그만큼 커진다**.
- 처음 띄울 때 화면이 비었다: Vite가 새 의존성으로 다시 최적화하느라 `504 (Outdated Optimize Dep)` — 새로고침으로 해결(일시적).
- **운영 배포**: 사용자가 `task deploy` → `[release] 끝: 636f269e0183`(헬스체크 2번째 통과, 마이그레이션 없음, 실행 파일 25MB). Claude가 바깥에서: 페이지 `200`, CSS에 고운바탕, 글꼴 조각 `200 font/woff2`. **사용자가 휴대폰으로 새 모양 확인.**

### 이번 세션 커밋
`4fe9181` 시안 A~C / `4bc5088` D / `9936589` D2 / `ca8efc0` D2 채택 / `fb5d60b` 로고 1~4 / `4d31ef9` 로고 5~9 / `1f335bc` 요즘 스타일 / `567dfbc` 로고 · 부제 확정 / (다음 두 커밋) 앱에 D2 반영 · 문서

### 다음에 할 일
- M4 체크인 설계(사전 검증 → 설계 · 의존성 먼저 보여 주기). 체크인 원도 그때.
- 목업 계획: S-04 사고 기록을 D2 결로 다시 그릴지(지금 `Record.*.html`은 옛 결) — 그 화면 구현(M5) 때.

---

## 2026-11-01 · M3 남은 완료 기준 3개 · M3 끝 (노트북, 호텔 와이파이)

### 한 일
- 2026-10-31 (5)의 "[검토 중] M3 확인을 미루고 M4 먼저"는 **정할 필요가 없어졌다** — 호텔 와이파이에서 SSH(22)가 열려 있어 남은 셋을 바로 했다(사용자 결정).
- **README대로 새 버전 배포**(SERVER 4.17절): Claude가 `task lint` · `task test`(종료 코드 0) · push(`f925ad6`) → 사용자가 `task deploy` → 서버 `f925ad636013`, releases에 옛 `1a04de8c6130`도 남음. 새 버전 내용은 Caddyfile에서 `Via` 헤더 빼기.
- **재부팅 자동 복구**(4.18절): 사용자가 `sudo reboot` → 1분 안에 `docker` · `cbt` · `caddy` · `cbt-backup.timer` 넷 다 active, Postgres · Valkey `Up 30 seconds (healthy)`, `NRestarts=0`. Claude가 바깥에서 `200` · health ok · `Via` 없음.
- **배포 실패 시 자동 되돌림**(4.19절): 서버에서 진짜 `release.sh`에 가짜 앱(버전은 맞게 대답, serve는 포트 안 엶) → 30초 health 실패 → `f925ad636013`으로 되돌림 → `헬스체크 통과 (2번째)` · `옛 버전으로 돌아옴`, 스크립트는 실패로 끝남(의도). `cbtctl version` `f925ad636013`, 가짜 폴더 지움, 바깥 health `200`.
- PLAN M3 ✅ — 완료 기준 7개 전부 확인. 회사 PC SSH 열쇠는 완료 기준이 아니라 열린 채로 둠.

### 막혔던 일
**`header_down -Via`로 `Via: 1.1 Caddy`가 안 지워짐**
- 증상: 새 Caddyfile을 넣고 reload했는데 `curl.exe -sI`에 `Via: 1.1 Caddy`가 그대로.
- 원인: `reverse_proxy`의 `header_down`은 앱(위쪽)이 보낸 응답 헤더를 다듬는 자리고, `Via`는 Caddy가 그 뒤에 붙인다. 같은 이유로 뺀 `-Server`는 사이트 `header` 블록에 있어서 지워졌다(그 블록의 지우기는 응답 직전에 적용).
- 해결: `-Via`를 `header` 블록으로 옮김. 이번엔 로컬 `caddy:2` 컨테이너에 프록시 · 위쪽 서버를 띄워 헤더가 실제로 지워지는 걸 본 뒤 서버에 넣음(`1406977`). ARCHITECTURE 7.2절에 반영.
- 배운 것: `caddy validate`는 문법만 본다. 설정의 **결과**(헤더)는 같은 프로그램을 띄워 응답으로 확인한다.

### 이번 세션 커밋
`f925ad6` Caddy `Via` 빼기(1차, 안 됨) / `1406977` `Via`는 `header` 블록에서 / (이 커밋) M3 끝 기록 — 서버 반영은 사용자가 침

### 다음에 할 일
- M4 체크인 — 위 "현재 상태"의 "다음에 할 일".
- 사용자가 정할 것: 새벽 자동 재부팅 · HSTS 1년(급하지 않음). 회사 PC SSH 열쇠는 회사 PC에서.

---

## 2026-10-31 (5) · 노트북 첫 작업 (노트북)

### 한 일
- 사용자가 노트북에서 clone · README "처음 한 번"까지 마침.
- Claude가 본 값: `main` 최신(`fc33d33`), Go 1.27.0 · Task 3.54.0 · pnpm 12.16.1 · Node 24.15.0 · golangci-lint 2.15.0(`ci.yml`과 같음) · Docker Desktop 29.9.2(WSL2), compose Postgres · Valkey healthy, 5432는 직접 설치한 PostgreSQL 18이 사용 중. `~/.ssh` 없음, `age` 없음.
- CLAUDE.md "기기" 표에 노트북 줄(개인 계정), "5432는 이미 차 있다"에 노트북 추가.
- **확인 못 한 것**: 노트북에서 `task test` · `task lint` · 로그인은 Claude가 돌리지 않았다.
- 노트북 열쇠 `laptop`을 Vultr 웹 콘솔(root)로 서버에 추가 → `ssh cbt`로 `sagming40@cbt-1` 접속 확인(첫 접속 지문 `yes`).
- 노트북 `age` 설치(사용자): `age --version` → `v1.3.2`.
- 로그인 화면의 `New release '26.04.1 LTS' available`은 따르지 않는다 — 24.04를 지원 종료(2029-04)까지 쓰고 새 서버로 옮기기로 했다(위 "기억할 날짜").

### 막혔던 일
**웹 콘솔을 못 찾음**
- 원인: 인스턴스 목록에는 없고 `cbt-1` 상세 화면 오른쪽 위 모니터 아이콘(또는 줄 끝 `···`)에 있다. 콘솔은 root라 `~`가 `/root` — 열쇠는 `/home/sagming40/.ssh/authorized_keys`에 넣어야 한다.
- 해결: SERVER 3.3 · 3.4절에 적음.

**`ssh cbt` → `Could not resolve hostname cbt: \354\225\214…`**
- 증상: 깨진 글자는 Windows 오류 문구 "알려진 호스트가 없습니다."의 UTF-8 바이트.
- 원인: 메모장이 `~/.ssh/config`를 `config.txt`로 저장해 ssh가 별명을 못 읽었다.
- 해결: 이름 바꿈, `ssh -G cbt`로 확인. SERVER 3.2절에 "파일 먼저 만들고 열기".
- 배운 것: ssh 설정이 읽혔는지는 접속 전에 `ssh -G 별명`으로 본다.

### [설계] 스왑: 1GB 만들기 → 있던 2.3GB 유지
- 증상: `sudo fallocate -l 1G /swapfile` → `Text file busy`. 바로 앞 `swapon --show`에 `/swapfile 2.3G`가 이미 있었다.
- 원인: Vultr가 서버를 만들 때 스왑 파일을 넣어 둔다. 2026-10-30 설계 때 접속 화면의 `Swap usage: 0%`를 "스왑 없음"으로 잘못 읽었다.
- 해결: 그대로 쓴다(권한 600 · fstab 등록 확인). `vm.swappiness` 60 → 10. ARCHITECTURE 7.5절 · SERVER 4.7절 · PLAN 고침.
- 배운 것: 서버를 바꾸는 명령 앞에 **지금 상태를 보는 명령**을 둔다 — 이번에는 그 덕에 바로 잡혔다. "0%"는 사용률이지 크기가 아니다.

### 이번 세션 커밋
`e9083c2` 기기 표 / `db2460e` 웹 콘솔 / `a3f12c7` config.txt 함정 / `997858a` 노트북 접속 / `404668b` age / `d52b8c5` 스왑 / `2a5a8aa` `cbt` 계정 · 폴더 / `5bc8098` 비밀값 · 운영 compose / `5ca01f3` DNS / `f7b60be` psql / `3c47714` Caddy / `f587856` 백업 열쇠 / `1a04de8` 운영 파일 · 백업 타이머 / (이 커밋) 첫 배포 — 전부 문서만, 동작 변경 없음

### 서버: 앱 전용 계정 `cbt` · 폴더
- 사용자가 침(SERVER 4.8절). 본 값: `uid=999(cbt) gid=987(cbt)`, 셸 `nologin`, `sudo su - cbt` 거부, `/opt/cbt` · `/etc/cbt` 755 · `/srv/cbt` 700 전부 root.

### 서버: 비밀값 · 운영 compose
- 사용자가 침(SERVER 4.9 · 4.10절). 비밀번호는 서버에서 `openssl rand -hex 32`로 만들어 화면에 띄우지 않고 두 파일에 넣음.
- 본 값: 비밀번호 길이 64 · 64, 두 파일 일치, 권한 `.env` 600 · `cbt.env` 640 root:cbt, `compose config` OK. Postgres · Valkey `Up (healthy)`, `PONG`, 로그 제한 `max-file:3 max-size:10m`, `ss`에서 127.0.0.1만 LISTEN, 노트북에서 `Test-NetConnection` 22 True · 5432 False · 6379 False(PLAN M3 완료 기준 "바깥에서 DB · Valkey 포트 접속 실패" 관찰).
- `psql … -c 'select version();'`이 아무것도 출력하지 않았다(오류도 없음) → `exec -T … -tAc 'show server_version;'`로 다시 봐서 `18.6 (Debian 18.6-1.pgdg13+2)`. TTY를 붙인 쪽의 원인은 따로 파지 않았다.
- Claude 실수 두 개(확인 명령): `grep -c '<'`가 주석 줄의 `<…>`까지 세서 1, `ls -l`이 `.env`를 숨김 → `grep -n` · `ls -la`로 다시 봐서 정상 확인.

### DNS
- 사용자가 Cloudflare에 `cbt` A 레코드(DNS 전용)를 만듦(SERVER 4.11절). 본 값: 목록에 `cbt.sagming.com A … DNS 전용 · 자동`, 노트북 `nslookup`이 1.1.1.1 · 지금 네트워크 DNS 둘 다 서버 IP.

### Caddy
- 사용자가 침(SERVER 4.12절): 공식 apt로 2.11.7, `deploy/Caddyfile` → `validate` `Valid configuration` → reload.
- Claude가 노트북에서 본 값: `https://cbt.sagming.com` → `502 Bad Gateway`(앱 없음, 인증서 오류 없음), `http://` → `308` `Location: https://cbt.sagming.com/`, 인증서 `CN=cbt.sagming.com` · Let's Encrypt `YE2` · TLS 1.3 · 만료 2027-01-29, `https://IP` → TLS 실패, `http://IP` → `https://IP`로 308.
- **확인 못 한 것**: 보안 헤더. 502 응답에는 HSTS 등이 없고 `Server: Caddy`도 남아 있었다 — 오류 응답은 Caddy가 처리 줄을 거치지 않고 새로 써서 `header` 설정이 안 붙는 것으로 보인다. 앱이 정상 응답할 때(첫 배포 뒤) 본다(PLAN M3에 추가).
- ARCHITECTURE 7.2절의 "`http://IP`는 빈 응답"은 실제와 달라서 고침(308 후 끊김, 앱은 여전히 안 보임).

### 인증서를 받자마자 스캐너가 들어옴
- 증상: reload 20초 뒤 `journalctl`에 해외 IP 한 곳에서 `/.env` · `/api/.env` · `/phpinfo.php` · `/wp-admin/setup-config.php` · `/database.sql` 요청이 몰렸다(전부 502).
- 원인: 새 인증서는 공개 목록(Certificate Transparency)에 올라가고, 봇이 그 목록을 보고 새 이름을 바로 훑는다. 우리를 노린 게 아니라 인터넷 소음이다.
- 대응: 없음 — 앱이 없었고, 비밀값은 저장소 · 웹 경로 밖(`/etc/cbt` · `/srv/cbt`)에 있다. 앱이 뜨면 기본 잠금 · 요청 제한이 받는다.
- 배운 것: 이름이 공개되는 순간부터 공격 대상이다 — "배포 전에 인증 먼저"(PLAN 1장)가 맞았다. Caddy는 접근 로그를 안 남겨도 오류 때는 IP · 경로를 journald에 남긴다(쿠키는 가림).

### 백업 열쇠
- 사용자가 침(SERVER 4.13절): 노트북 `age-keygen` → 서버 `apt install age`(1.1.1) · `/etc/cbt/backup.pub`. 짝 시험: 서버에서 잠근 파일을 노트북 개인키로 열어 `열쇠 짝 시험` 출력.
- 보관: 비밀번호 관리자가 없어서 노트북 파일 + 종이로 정함. 집 · 회사 PC에는 나중에 USB로(README "운영 서버를 다루는 PC라면" 3).

### 운영 파일 · 백업 타이머
- 사용자가 침(SERVER 4.14절). `verify`는 `cbt.service`의 실행 파일 없음만 냄(첫 배포 전이라 정상).
- 본 값: 손으로 돌린 백업 `[backup] 만듦 … (1096 바이트)`, 타이머 NEXT `2026-10-31 19:09:36 UTC`(= 04:09 KST), 노트북 `task backup:pull`로 같은 파일 수신.
- Claude가 노트북에서: 받은 백업을 개인키로 풀어 896바이트 · 머리글 `PGDMP` — **실제 Postgres의 `pg_dump` → `age` → 가져오기 → 풀기**가 처음으로 끝까지 됐다(2026-10-31 (3)의 "확인 못 한 것" 중 `pg_dump`).
- Claude가 노트북에서 `task deploy -- -DryRun`: 프론트 빌드 → embed → `bin/app-linux-amd64`(`main.version=f587856d1fd0`) → 보낼 폴더까지 성공, 작업 트리 깨끗.

### 첫 배포
- 사용자가 `task deploy`(커밋 `1a04de8`) → `sudo cbtctl user create`(SERVER 4.15절). `cbt` active, `verify` 출력 없음.
- Claude가 노트북에서: `https://cbt.sagming.com` `200` + 보안 헤더 4개 · `Server` 없음(Caddy 단계에서 미룬 확인 끝), `/api/health` postgres · valkey ok, 비로그인 `/api/me` `401`.
- 사용자 화면: 노트북 브라우저 · 휴대폰(LTE) 로그인 → 오늘 · 설정 → 로그아웃 둘 다 됨. 휴대폰 오늘 화면 서버 상태 Postgres · Valkey `ok`. — **PLAN M3 완료 기준 1(휴대폰 · PC 로그인)**, 단 "회사 PC"가 아니라 노트북으로 봤다.
- 남은 점: 응답에 `Via: 1.1 Caddy`(프로그램 이름을 숨기려던 `-Server`와 어긋남, 보안 영향은 작음).

### 업로드 규칙 예외 (사용자 결정)
- 사용자가 기기별 전역 설정(`~/.claude/CLAUDE.md`) 사본을 구글 드라이브 `_private/.claude/`에 모아 통일하는 중. 사본에 이 프로젝트 이야기(배포 주소 · 서버 사양)가 들어가 CLAUDE.md "업로드 금지"와 어긋났다 → 그 폴더만 예외로(비밀값 · 실제 기록은 여전히 금지). CLAUDE.md "업로드 금지 · 데이터 취급".
- 노트북 전역 설정은 집 PC 것(말투 · 공용 인프라)을 합치고 오늘 바뀐 것(첫 배포 · `laptop` 열쇠)을 더해 다시 씀, 드라이브 사본과 같은 내용. 집 PC 원본의 공용 인프라 갱신은 집 PC에서.

### 완료 기준: 요청 제한이 실제 접속 IP 기준인가
- Claude가 노트북에서 운영 `/api/auth/login`에 없는 이메일(`…@example.invalid`)로 틀린 로그인을 보냄. 진짜 계정은 안 건드림.
- 같은 이메일 21번 → 1~5번 `401`, 6번부터 `429` + `Retry-After: 894` — **계정 제한**(실패 5번)이 먼저 걸렸다. IP 제한을 보려면 이메일을 매번 바꿔야 했다.
- 이메일을 바꿔 25번 더 → 전부 `401`. 20번이 넘었는데 안 막혀서 서버 Valkey를 봄(사용자가 침, IP 뒤 두 자리 가림): `rl:login:ip:223.49.x.x 횟수=20` · `rl:login:ip:223.48.x.x 횟수=11` — **노트북의 바깥 IP가 둘**이라 요청이 나뉘었다(이동통신사 대역 — 통신사가 나가는 주소를 돌려 씀). `127.0.0.1` 키는 없음 = 앱이 Caddy 주소가 아니라 **실제 접속 IP**로 센다.
- 다시 이메일을 바꿔 보냄 → 35번째에 `429`(이메일이 매번 달라 계정 제한일 수 없음 = **IP 제한**), 바로 다음 요청은 다른 IP로 나가 `401`. 한 IP가 막혀도 다른 IP는 안 막힘 = "서로의 제한에 걸리지 않음"을 직접 봤다. (중간 1번은 연결 실패 `000`)
- **PLAN M3 완료 기준 3 확인.** 휴대폰으로 따로 해 보지는 않았다 — 두 IP가 각자 세지는 것으로 갈음.
- 배운 것: 같은 기기라도 바깥 IP가 하나라는 보장이 없다(통신사 NAT). 시험 결과가 설계와 다르면 코드보다 먼저 **서버에 실제로 쌓인 값**을 본다.

### 완료 기준: 복원 연습
- KTX 공용 와이파이라 서버 접속(passphrase)이 어려워, 사용자가 받아 둔 최신 백업 `cbt-20261031T102225Z.dump.age`(6902바이트, 계정 생성 뒤)로 노트북 안에서 Claude가 함(SERVER 4.16절).
- 본 값: 풀어서 6702바이트 → 로컬 빈 DB `cbt_restore_test`에 `pg_restore` 종료 코드 0 → 테이블 `goose_db_version` · `sessions` · `users`, `users` 1건, 마이그레이션 버전 2(저장소와 같음). 치운 뒤 컨테이너 안 파일 0 · 시험 DB 0 · 임시 파일 없음.
- 운영 건수는 서버에서 세지 않았다 — 가입이 꺼져 있고 `cbtctl`로 만든 계정 하나뿐이라 1로 갈음. 다음 서버 접속 때 한 줄로 확인(SERVER 4.16절).
- PowerShell 5.1에서 덤프를 `|` 파이프로 컨테이너에 넣으면 바이너리가 깨진다 → `docker compose cp`로 복사해서 넣음.
- README "상태" 갱신, "배포" 절 추가(push → `task deploy` → 확인, 자동 되돌림, DryRun, 백업 · 복원 위치).

### [검토 중 — 확정 아님] M3 남은 확인을 미루고 M4 먼저 할까
> 사용자가 물어서 Claude가 낸 안이다. **아직 정하지 않았다** — PLAN은 고치지 않았다. 다음 세션에서 사용자와 정한다.
- 안: M3의 "재부팅 자동 복구" · "배포 실패 시 자동 되돌림" 확인을 **M6 시작 전 필수 항목**으로 옮기고, M3는 "배포 완료, 확인 2개는 M6 전으로"로 닫은 뒤 M4로 간다. 정하면 PLAN M3 · M6과 2장 표를 고치고 여기 이유를 옮겨 적는다.
- 근거:
  1. 서버에 지킬 데이터가 없다 — 실제 기록은 M6부터(PLAN 1장). 지금은 계정 하나뿐이라 재부팅 · 배포가 꼬여도 잃는 것이 없다. 두 확인이 진짜 중요해지는 때는 M6 직전이다.
  2. 남은 셋 중 "README대로 새 버전 올리기"는 M4 완료 기준(배포 서버에서 휴대폰 체크인)을 하면서 저절로 채워진다.
  3. 놔둬도 저절로 터질 일이 드물다 — 새벽 자동 재부팅은 꺼 뒀고(ARCHITECTURE 7.5절), 되돌림은 컨테이너에서 시나리오 10개로 돌려 봤다(2026-10-31 (2)). 실제 서버에서만 안 해 봤다.
- 덧붙인 안: M4를 처음 배포할 때 안정된 네트워크라면 그 자리에서 두 확인도 해 버린다(합쳐 10분 남짓).
- M4를 시작하면(정해지면): ① REQUIREMENTS C1 · C2 · C4, DATA_MODEL, DESIGN S-02 · S-03, 목업을 읽고 사전 검증(특히 시간대 경계 — 자정 근처 체크인이 어느 "오늘"인가, 남의 데이터 404) → ② 설계 · 보완점 · 진행 순서 · 새 의존성을 사용자에게 먼저 보여 줌 → ③ 마이그레이션 → API · 테스트 → 화면 → 배포. 직접 칠 부분은 그때 정한다(학습 포인트: `EmotionPicker` 화면 부품, 시간대 계산).

### 다음에 할 일
- 먼저 위 "[검토 중]"을 사용자와 정한다(M4 먼저 / M3 확인 먼저).
- M3 완료 기준 남은 셋(서버 접속 필요, 안정된 네트워크에서): README "배포" 순서대로 새 버전 한 번(`Via` 헤더 빼기를 겸해도 좋음) → 서버 재부팅 뒤 앱 · DB · Caddy 자동 복구 · health 200 → 배포 중 health 실패를 일부러 내서 자동 되돌림 확인. 운영 `users` 건수 한 줄도 그때.
- 선택: Caddyfile `reverse_proxy`에 `header_down -Via`
---

## 2026-10-31 (4) · 세 번째 기기(노트북) 준비 문서 (집 PC)

### 한 일
- 사용자가 노트북에서도 작업하기로 함(아직 clone 전). 어느 기기에서든 바로 이어 가도록 문서 정리.
- CLAUDE.md "기기 둘" → "기기": 노트북 줄 추가(계정 · 도구는 처음 작업할 때 채움), 새 기기 시작 순서.
- README "처음 한 번"에 0단계 `git clone` · 끝났는지 확인하는 법, "운영 서버를 다루는 PC라면" 절(SSH 열쇠 · ssh 별명 `cbt` · `age` — `winget show FiloSottile.age`로 id 확인, 1.3.2).
- "두 PC"라고 적힌 곳 중 기기 수에 기대던 문장을 고침(ARCHITECTURE 7.3 · 7.4절, PLAN M3, SERVER 3.4절). 회사 · 집 PC의 사실(5432 사용 중 등)은 그대로.
- DEVLOG "현재 상태"에 "서버에 접속되는 기기"(지금은 집 PC뿐)와 기기별 할 일.

### 이번 세션 커밋
(이 커밋) 문서만 — 동작 변경 없음

### 다음에 할 일
- 위 "현재 상태"의 "다음에 할 일" · "기기별로 할 일". 다른 기기에서 서버 작업부터 하려면 그 기기 열쇠를 먼저 서버에 넣는다.

---

## 2026-10-31 (3) · M3 백업 스크립트 · 타이머 (집 PC)

### 한 일
- 타이핑 7단계(사용자가 치고 Claude 검수): `deploy/cbt-backup` · `cbt-backup.service` · `cbt-backup.timer` · `task backup:pull`. 설계 먼저: ARCHITECTURE 7.4절 세부.
- 확인한 값: 사용자가 친 파일이 시안과 같음(탭 들여쓰기), `bash -n` · ShellCheck 경고 0. Ubuntu 24.04 컨테이너(age 1.1.1, docker는 가짜)에서 공개키 없음 거부 · 정상 백업을 개인키로 풀면 원본 · 공개키로는 못 품 · 덤프 실패 시 파일 없음 · 같은 초 거부 · 20번 뒤 14개 · 비root 거부, `systemd-analyze verify` exit 0, 타이머 다음 실행 19:00 UTC(= 04:00 KST).
- **확인 못 한 것**: 실제 Postgres의 `pg_dump` · 서버 `scp`. 서버 준비 뒤 복원 연습에서 본다.

### [설계] 시안을 돌려 보고 바꾼 것
- **크기로 빈 덤프를 가리려던 검사를 뺐다**: 빈 입력도 `age`로 잠그면 200바이트쯤 돼 검사가 늘 통과했다. `pg_dump` 종료 코드(pipefail)로 판단하고, 쓸 만한지는 복원 연습으로 확인한다.
- **같은 초에 두 번 돌면 거부**: 이름이 초 단위 시각이라, 시험에서 실패한 덤프가 바로 앞의 좋은 백업을 덮어썼다(배포 직전 백업과 타이머가 겹치면 실제로 생길 수 있다).
- 배운 것: "검사가 있다"와 "검사가 실패를 잡는다"는 다르다 — 일부러 실패를 넣어 검사가 우는지 본다.

### 이번 세션 커밋
`b4b4b42` 백업 설계 / `ee44983` 백업 스크립트 · 타이머 · `task backup:pull` / (이 커밋) 문서 반영

### 다음에 할 일
- 서버 작업(Claude가 명령과 이유를 보여 주고 사용자가 친다, SERVER.md에 기록): 스왑 1GB · `cbt` 계정 · 폴더 · 비밀값(`openssl rand -hex 32`) · 운영 compose · Caddy(공식 apt) · `age` 설치 · 열쇠(개인키는 PC에 보관) · 운영 파일 설치
- PC: `age` 설치(복원용), `~/.ssh/config`에 별명 `cbt`

---

## 2026-10-31 (2) · M3 운영 파일 2 ~ 6단계 (집 PC)

### 한 일
- 타이핑(사용자가 치고 Claude 검수): 2 운영 compose · env 예시, 3 Caddyfile, 4 `cbt.service` · `cbtctl`, 5 서버 쪽 배포 `release.sh`.
- 자동(사용자 지시): 6 `scripts/deploy.ps1` · `task deploy`.
- 설계 문서 먼저: ARCHITECTURE 7.2절(Caddy 인증서 이메일 안 넣음), 7.3절(서버 쪽 · PC 쪽 배포 세부), SERVER 3.2절(ssh 별명 `cbt`가 필요).
- 확인한 값:
  - Caddyfile: `caddy validate` → `Valid configuration`, `caddy fmt --diff` 0줄.
  - `cbt.service`: Ubuntu 24.04 컨테이너에서 `systemd-analyze verify` exit 0, `security --offline` 4.5 OK.
  - `cbtctl` · `release.sh`: `bash -n` · ShellCheck 경고 0.
  - `release.sh`: 컨테이너에서 시나리오 10개(첫 배포 · 백업 스크립트 없음 거부 · 정상 · 헬스체크 실패 되돌림 · migrate 실패 · 버전 불일치 · `-dirty` 거부 · 같은 커밋 재배포 · 백업 실패 · 동시 배포) + 6번 배포 뒤 5개만 남음. `systemctl` · `curl`은 가짜.
  - `task deploy`: 커밋 안 한 변경 → 거부, push 안 한 커밋 → 거부, `-DryRun` → 보낼 폴더에 `ELF` 실행 파일(컨테이너에서 `version` = 커밋 번호) + LF `release.sh`, 빌드 뒤 작업 트리 깨끗.
- **확인 못 한 것**: 실제 서버에서의 `scp` · `ssh -t` · `sudo` · systemd 재시작 · 헬스체크. 서버 준비 뒤 첫 배포에서 본다.

### [설계] 정하면서 바꾼 것
- Caddy 인증서 이메일을 안 넣는다: Let's Encrypt가 2025년에 만료 알림 메일을 없애서 쓸 곳이 없다(어제 계획한 `ACME_EMAIL` env 철회, ARCHITECTURE 7.2절).
- `release.sh`(시나리오를 돌려 찾음): 백업 스크립트가 아직 없을 때 첫 배포만 건너뛴다(DB가 비어 있음), 실행 파일을 `app.new`로 넣고 버전 확인 뒤 바꾼다, 링크는 `mv`로 바꿔치기, 동시 배포 잠금(`flock`).
- `deploy.ps1`: 서버에서 임시 폴더를 만들면 ssh가 한 번 더 들어서, PC에서 폴더를 만들어 `scp -r` 한 번으로 보낸다. 서버 주소는 ssh 별명 — 저장소에 IP를 적지 않는다.

### 막혔던 일
**Caddyfile 검사가 Git Bash에서 "파일 없음"으로 실패했는데 "통과"로 커밋할 뻔함**
- 원인: Git Bash(MSYS)가 `/etc/caddy/Caddyfile` 같은 인자를 Windows 경로로 바꿔 Docker에 넘겼다.
- 해결: `MSYS_NO_PATHCONV=1`로 다시 돌려 `Valid configuration` 확인. 커밋 내용은 맞았다.
- 배운 것: 검사 명령은 **출력까지 읽고** 나서 "통과"라고 쓴다.

**`bash -n .\deploy\cbtctl` → `.deploycbtctl: No such file`**
- 원인: PowerShell에서 bash에 백슬래시 경로를 넘기면 bash가 `\`를 이스케이프로 먹는다.
- 해결: bash에 주는 경로는 `deploy/cbtctl`(슬래시).

**`release.sh` 들여쓰기에 탭과 공백이 섞임**
- 증상: 화면에서는 똑바로 보이는데 파일에는 `··→`처럼 섞여 있었다. "Convert Indentation to Tabs"는 4칸만 탭으로 바꾸고 2칸은 남겼다.
- 원인: 에디터가 `Spaces: 2` 설정이라 Tab 키가 공백을 넣고, 탭도 2칸 폭으로 보여 줘서 눈으로 구분이 안 됐다.
- 해결: 공백만 Claude가 고쳐 커밋(사용자 지시). 에디터를 "Indent Using Tabs"로 바꾸기를 권함.
- 배운 것: 들여쓰기는 눈이 아니라 `cat -A`(탭 = `^I`)로 본다.

**PowerShell 5.1은 BOM 없는 `.ps1`의 한글을 깨뜨린다**
- 원인: Windows PowerShell 5.1은 BOM이 없으면 파일을 시스템 코드 페이지(CP949)로 읽는다.
- 해결: `scripts/deploy.ps1`을 UTF-8(BOM) + CRLF로 저장(`.gitattributes`가 `*.ps1`을 CRLF로 지킨다). 실행해서 한글 출력 확인.

### 이번 세션 커밋
`cb5164c` compose · env 예시 / `494949f` Caddy 이메일 안 넣기 / `d7db848` Caddyfile / `0b9c06c` `cbt.service` · `cbtctl` / `f4523a1` · `67192ec` `release.sh` 설계 · 파일 / `8ecf23b` · `dbcd689` `deploy.ps1` 설계 · 파일 / (이 커밋) 문서 반영

### 다음에 할 일
- 7단계: 백업 스크립트(`pg_dump` → `age` → `/var/backups/cbt`, 14개) · systemd 타이머 · `task backup:pull` — `release.sh`가 부르는 `cbt-backup`
- 그다음 서버 작업(SERVER.md에 적으며): 스왑 · `cbt` 계정 · 폴더 · 비밀값 · compose · Caddy · `age` 열쇠

---

## 2026-10-31 · M3 운영 구성 결정 · 시나리오 사전 검증 (집 PC)

### 한 일
- 운영 구성 결정 3개를 추천안대로 확정(사용자): ① 실행 파일 + systemd(DB · Valkey만 compose) ② PC에서 `task deploy` ③ `age`로 암호화한 백업을 PC로 가져오기. 새벽 자동 재부팅은 처음엔 끈다.
- 설계 문서 먼저: ARCHITECTURE 7절을 "빌드 · 배포 · 운영"으로 넓힘, PLAN M3 작업 · 완료 기준 갱신.
- 작업 방식: 이번 운영 파일은 **타이핑 학습 모드**(사용자 요청). 설계 문서만 Claude가 자동으로 고친다.

### [설계] 사전 검증을 "시나리오까지" 다시 돌림
- 어제 찾은 6개(스왑 · Docker 로그 · IP 직접 접속 · HSTS · CSP · 전용 계정)는 구성을 정적으로 훑은 것이었고, **흐름을 끝까지 돌려 보지 않았다**(사용자가 물어 다시 함). 시나리오별로 새로 찾은 것:
  1. 첫 배포에 DNS 전에 Caddy를 켜면 인증서 발급 실패가 반복돼 Let's Encrypt가 막는다 → DNS 먼저 · `nslookup` 확인
  2. 재부팅 때 앱이 DB보다 먼저 떠 죽는다 → systemd `Restart=always`, compose `restart: unless-stopped`, 완료 기준에 "재부팅 후 자동 복구"
  3. 마이그레이션 성공 뒤 새 버전이 고장 나 되돌리면 옛 버전이 새 스키마에서 깨질 수 있다 → 옛 버전과 맞는 마이그레이션 규칙 + 배포 직전 자동 백업
  4. PC 배포가 passphrase를 3 ~ 4번 묻는다 → 복사 1번 + 원격 실행 1번으로 묶음
  5. `age` 개인키를 잃으면 백업 전부를 못 연다 → 만들자마자 따로 보관을 필수 단계로
  6. 서버에서 운영 env로 `app user create`를 돌리는 명령이 길다 → `cbtctl` 도우미
  7. Valkey에 비밀번호를 걸면 compose 상태 확인(`ping`)이 실패한다 → 확인에도 비밀번호
  8. Caddyfile에 이메일을 쓰면 공개 저장소에 올라간다 → `ACME_EMAIL` env
  9. release 폴더가 쌓인다 → 5개만 남김
  10. 자동 업데이트는 커널을 설치만 하고 재부팅하지 않는다 → 자동 재부팅은 2번을 확인한 뒤 판단
- 다시 돌려 봐도 문제없던 것: Caddy가 `X-Forwarded-For`를 새로 붙여 `TRUSTED_PROXIES=127.0.0.1`로 실제 IP가 나옴, `Host` 그대로라 CSRF 차단 · `Secure` 쿠키 동작, 순수 Go라 Windows에서 Linux용 빌드, push된 커밋만 배포해 두 PC가 덮어쓰지 않음.
- 배운 것: 사전 검증 보고에는 **어떤 기준 · 어떤 시나리오까지 돌렸는지**를 같이 적는다.

### 타이핑 1단계: `app version` · `task build:linux`
- 사용자가 치고 Claude가 검수. 확인: `go run ./cmd/app version` → `dev`, `bin/app-linux-amd64` 첫 4바이트 `ELF`, 버전 문자열 `<커밋>-dirty` 포함, `task lint` 0 issues · `task test` 통과.

### 막혔던 일
**`-X flag requires argument of the form importpath.name=value`**
- 원인: `-X main.version{{.COMMIT}}` — `=`를 빠뜨렸다(타이핑). 같은 로그에서 버전이 `v1-legacy-56-g…`로 나온 것도 발견 — `git describe`가 가장 가까운 태그(`v1-legacy`)를 앞에 붙였다(Claude가 준 코드의 실수).
- 해결: `=` 추가, `git describe … --exclude="*"`(태그를 안 쓰고 커밋 번호만).

**빌드는 성공했는데 `bin/app-linux-amd64`가 Windows 실행 파일**
- 증상: 파일 첫 바이트가 `MZ`(Windows .exe 표시). Linux 실행 파일이면 `ELF`.
- 원인: Claude가 준 코드에서 `GOOS · GOARCH · CGO_ENABLED`를 명령(`cmd`) 하나에 `env:`로 걸었는데, Task는 그 자리의 env를 **오류 없이 무시**한다. GOOS가 안 들어가 Go가 지금 PC(Windows)용으로 만들었다.
- 해결: `env`를 작업(task) 단위로 올림. 같은 작업의 `web:build` · `embed`는 GOOS를 안 봐서 영향 없음.
- 배운 것: "빌드 성공"은 결과물이 맞다는 뜻이 아니다. **만들어진 파일 자체를 연다**(첫 바이트 · 버전 문자열). 검수에서 코드만 읽었으면 놓쳤다.

### 이번 세션 커밋
`bc32fc1` 설계 문서 / (직전 커밋) `app version` · `task build:linux` / (이 커밋) 문서 반영

### 다음에 할 일
- 타이핑 2단계: `deploy/compose.prod.yaml` + env 예시

---

## 2026-10-30 · M3 운영 구성 설계안 (결정 대기)

> 대화에서 사용자에게 보여 준 설계안이다. 결정이 나면 ARCHITECTURE 7절 · SERVER.md · PLAN에 옮기고 이 블록은 그대로 둔다.

### 구성
```
바깥 ─ 443/80 ─▶ Caddy(인증서 자동 · 보안 헤더) ─ 127.0.0.1:8080 ─▶ app serve(systemd, 전용 계정 cbt)
                                                        └─ 127.0.0.1만 ─▶ docker compose: Postgres · Valkey(비밀번호)
```
- 서버 위치: `/opt/cbt/releases/<커밋>/app` + `/opt/cbt/current`(지금 버전 링크, 되돌리기 쉬움), 비밀값 `/etc/cbt/cbt.env`(600), 데이터 `/srv/cbt/compose.yaml`
- 운영 env: `APP_ENV=production` · `HTTP_ADDR=127.0.0.1:8080` · `TRUSTED_PROXIES=127.0.0.1` · `ALLOW_SIGNUP=false` · DB · Valkey 주소(비밀번호는 서버에서 무작위 생성)

### 결정할 것 (추천)
1. **앱 실행**: 실행 파일 + systemd (추천) / 전부 Docker compose — 0002 "실행 파일 하나"와 맞고, 이미지 레지스트리 · 서버 빌드가 필요 없다. DB · Valkey만 compose.
2. **배포**: PC에서 `task deploy` (추천) / GitHub Actions 자동 — 서버 열쇠를 GitHub에 맡기지 않는다. 스크립트는 작업 트리가 깨끗하고 push된 상태가 아니면 거부, 실행 파일에 커밋 번호를 새긴다. 흐름: Linux용 빌드 → 복사 → migrate → 재시작 → health 확인.
3. **서버 밖 백업**: 매일 `pg_dump` → `age`로 암호화(공개키만 서버, 개인키는 PC) → `task backup:pull`로 PC에 가져오기 (추천) / Cloudflare R2.

### 사전 검증에서 넣은 것
- 스왑 1GB(지금 0 — 메모리 1GB에 넷이 돌다 업데이트 때 강제 종료 위험)
- Docker 로그 크기 제한(기본은 무제한으로 쌓여 디스크가 찬다)
- IP로 직접 접속하면 앱을 보여 주지 않음(`cbt.sagming.com`으로만 응답)
- HSTS는 하루짜리로 시작 → 확인 뒤 1년
- CSP는 보류(SvelteKit 인라인 스크립트 해시 필요, M4 이후). nosniff · 프레임 금지 · Referrer는 지금
- 앱 전용 로그인 불가 계정 `cbt` + systemd 보호 옵션
- 그대로 되는 것: Caddy가 Host · X-Forwarded-For를 넘겨 CSRF 차단 · 실제 IP 규칙이 동작

### 진행 순서
1. Claude(자동): 설계 문서 → `deploy/`(compose · Caddyfile · systemd 유닛 · 운영 env 예시 · 백업 스크립트) → Taskfile(`build:linux` · `deploy` · `backup:pull`) → 로컬 검증
2. 사용자(서버): 스왑 · `cbt` 계정 · 폴더 · 비밀값 · compose · Caddy 설치
3. 사용자(Cloudflare): `cbt` A 레코드, 회색 구름
4. 같이: 첫 `task deploy` → 서버에서 내 계정 만들기
5. 같이: 완료 기준(PLAN M3) 확인

### 같은 날 한 일
- 서버 명령을 정리한 런북 [SERVER.md](SERVER.md) 새로 만듦(`78d1f26`).

### 다음에 할 일
- 위 결정 3개 → 진행 순서 1부터. 회사 PC에서 이어 하면 먼저 [SERVER.md](SERVER.md) 3.4절로 회사 PC 열쇠를 서버에 추가(집 PC로 접속해 둔 상태가 아니면, 회사에서 Vultr 웹 콘솔 비상구로 넣거나 집에서 미리 추가).

---

## 2026-10-30 · M3 서버 만들기 · 기본 보안 (집 PC, 사용자 직접)

### 한 일
- SSH 키: PC마다 따로 만든다(한 PC를 잃으면 그 키만 지운다). 집 PC `~/.ssh/cbt_vultr`(ed25519, passphrase 있음), Vultr에 `home-pc`로 등록.
- Vultr 서버 `cbt-1`: Shared CPU `vc2-1c-1gb`(1 vCPU · 1GB · 25GB SSD, $5/월) · 서울 · Ubuntu 24.04 LTS · **IPv4만** · 자동 백업 켬(+$1/월, 11-30에 다시 판단). 접속 사용자 `sagming40`(sudo).
- 기본 보안 (확인한 관찰값):
  - `cloud-init status` done → `apt upgrade` 65개 → 재부팅 → `uname -r` = 6.8.0-147-generic
  - ufw: 기본 deny incoming · allow outgoing, 열린 문 22 · 80 · 443뿐(Vultr가 22를 미리 열어 둠, 다른 규칙 없음)
  - SSH 잠금 `/etc/ssh/sshd_config.d/00-hardening.conf`(PermitRootLogin no · PasswordAuthentication no · KbdInteractiveAuthentication no) → 바깥에서: root 접속 `Permission denied (publickey)`, 비밀번호로 시도해도 묻지 않고 거부, `sagming40` + 키는 접속됨

### [설계] 정하면서 바꾼 것
- **OS는 24.04 LTS** (사용자 결정). 26.04 LTS(지원 2031)도 후보였지만 자료가 많은 쪽. 지원 종료 2029-04 → 그 전에 새 서버로 복원해 옮긴다. 접속 화면의 `do-release-upgrade` 안내는 따르지 않는다.
- **자동 백업은 켠다**: 처음엔 "서버 값의 20%"라 끄자고 했지만 실제 금액은 월 $1이고, 서버 설정을 처음 해 보는 지금은 "서버 통째로 되돌리기"가 값어치가 있다. 우리 DB 백업이 생긴 뒤 11-30에 다시 판단.
- **IPv6는 끈다**: 문이 하나 더 생겨 방화벽 · DNS를 두 벌 관리해야 하고, **요청 제한이 IP 단위라 IPv6(사용자 한 명이 /64 블록을 가짐)에서는 주소를 바꿔 가며 피할 수 있다**. IPv6를 켤 때는 요청 제한 키를 IPv6 /64 단위로 묶도록 코드를 먼저 고친다(지금 코드의 구멍 — 메모).
- 업데이트 중 "이미 수정된 설정 파일을 덮어쓸까?"(sshd_config · PAM)는 전부 **지금 것 유지**. Vultr가 맞춰 둔 상태가 정상이고, SSH는 `sshd_config.d/00-…` 파일로 덮어쓴다(sshd는 **먼저 읽은 값**을 쓰므로 Vultr의 `50-cloud-init.conf`보다 앞 번호).

### 막혔던 일
**`ssh-keygen`이 `No such file or directory`**
- 원인: `~/.ssh` 폴더가 없었다. `-f`로 위치를 정하면 폴더를 만들어 주지 않는다.
- 해결: `New-Item -ItemType Directory -Force "$HOME.ssh"` 뒤 다시.

**서버 안에서 `ssh`를 쳐서 서버가 자기 자신에게 접속하려 함**
- 원인: 새 계정 접속 테스트를 PC의 새 탭이 아니라 root 탭(서버 안)에서 쳤다. Windows 경로(`$HOME.ssh…`)가 리눅스에서 안 맞아 `Identity file … not accessible`.
- 배운 것: 명령 치기 전에 프롬프트를 본다 — `PS C:…>` = 내 PC, `root@cbt-1:~#` · `sagming40@cbt-1:~$` = 서버.

**`docker run hello-world`가 `failed to connect to the docker API … no such file`**
- 원인: 설치 명령을 옮기다 맨 앞 `docker-ce`(Engine 본체)가 빠지고 `docker-ce-cli`(리모컨)부터 설치됐다. 본체가 없으니 명령을 받을 곳이 없다.
- 해결: `sudo apt-get install -y docker-ce` → `Hello from Docker!`. 자동 보안 업데이트도 켜져 있음을 확인.
- 배운 것: 설치 뒤 "설치된 패키지 목록"을 계획과 한 번 맞춰 본다(이번엔 4개 = 하나 빠짐).

### 이번 세션 커밋
`60acc9c` 서버 · 기본 보안 기록 / (이 커밋) Docker 설치 기록

### 다음에 할 일
- 서버: 자동 보안 업데이트 확인 → Docker → 운영용 compose(Postgres · Valkey, 127.0.0.1에만) → Caddy → 앱 → DNS

---

## 2026-10-30 · M3 시작 ─ 배포 환경 결정 · Vultr · 도메인 (집 PC)

### 한 일
- 다른 프로젝트들이 배포에서 막힌 이유를 이 PC의 세션 기록 · 문서에서 찾았다(사용자가 준 claude.ai 웹 세션 링크는 이 앱에서 열리지 않음): AWS 새 가입 방식 "프로젝트" 계정이 조직 SCP로 시드니만 허용, Route 53 도메인 등록 실패, 가입이 덜 끝난 예전 일반 계정 → Support 케이스 답 대기. **기술이 아니라 계정 문제**였다.
- 개념 설명: 서버(VPS) · 도메인/DNS · HTTPS 인증서 · 리버스 프록시(Caddy) · 방화벽.
- 시드니/서울 왕복을 실측(0007 표) → 처음 여는 속도 때문에 **Vultr 서울**로 결정 (결정 기록 0007, 사용자).
- 사용자 직접: Vultr 가입(Google) · 카드 연결(`$0.00 deposit`) · 크레딧 $250 확인(만료 2026-11-30), Cloudflare 가입 · 2단계 인증(앱 + 이메일) · 백업 코드, `sagming.com` 등록(연 $10.46, 만료 2027-10-30, 자동 갱신, 등록자 정보는 WHOIS에서 가려짐).

### [설계] 0.13초는 잰 값이 아니었다
- 다른 프로젝트 문서에 적힌 "시드니 왕복 0.13초"를 근거로 처음엔 시드니를 추천했다. 실제로 재 보니 집 PC에서 약 0.23초(새 연결 첫 응답 약 0.7초)였다. 숫자를 근거로 결정할 때는 **출처가 측정인지 추정인지**부터 확인한다.

### 막혔던 일
**Vultr 가입 중 프로모션을 지워 버림**
- 증상: 카드 등록 화면에서 $250 프로모션의 `Remove`를 눌러 버렸고, 처음부터 다시 가입하려니 "That Google account is already in use".
- 원인: Google로 가입한 순간 계정은 이미 생겼다. 카드 등록 단계에서 멈춘 것뿐.
- 해결: 로그인 → My Org → Billing → **Make a Payment** 화면을 다시 여니 프로모션(`250VULTRFLY`)이 다시 붙어 있었다. 카드만 연결(`$0.00 deposit`)하니 크레딧이 들어왔다.
- 배운 것: 다른 이메일로 계정을 하나 더 만들어 혜택을 다시 받는 건 약관 위반(계정 정지 위험). AWS에서 계정이 둘로 갈라져 고생한 것과 같은 함정이라 피했다.

**결제 화면 캡처에 카드 유효기간 · CVV가 그대로 찍힘**
- 카드 번호는 가렸지만 유효기간 · CVV가 보였다. 번호 전체가 없어서 위험은 작지만, 앞으로 결제 · 개인정보 화면은 **번호 · 유효기간 · CVV · 주소 · 전화를 전부 가리고** 캡처한다. 기록이 남는 대화이기 때문이다.

### 이번 세션 커밋
(이 커밋) 결정 기록 0007 · M3 진행 반영

### 다음에 할 일
- SSH 키(사용자 직접) → Vultr 서버 만들기 → 서버 기본 보안 → Caddy · 앱 · compose → DNS

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

---

## 2026-10-30 · M1 시작 ─ 도구 설치 · v1 삭제 · compose (회사 PC)

### 한 일
- 개념 설명: Go(패키지 · 에러 값 · 인터페이스), Valkey와 Redis의 관계, SvelteKit SPA · SPA fallback, goose.
- 회사 PC에 winget으로 Go 1.27.0 · pnpm 12.15.0 · Task 3.54.0 설치 (사용자 직접).
- 타이핑 학습 모드로 v1 코드 삭제 · `.gitignore` 재작성 · `.gitattributes` · `compose.yaml` · `.env.example` 작성.
- 확인: 두 컨테이너 healthy, PostgreSQL 18.6 · Valkey 9.1.3, 포트는 `127.0.0.1`에만 열림, AOF 켬 — TTL 600초 키가 Valkey 재시작 뒤에도 남음(ttl 595).
- 질문 정리: "Alembic을 쓸 수 있나" → 쓸 수는 있지만 Python 런타임이 필요해 실행 파일 하나 배포가 깨지고, sqlc가 `.py` 마이그레이션을 못 읽어 스키마 원본이 둘이 된다. goose 유지.
- 외부 API 후보 검토: 기록 내용을 보내지 않는 것만 고름 → REQUIREMENTS A10(유출 비밀번호 거절 — L로 적었다가 사용자 결정으로 M, PLAN M2에 추가) · X3(날씨) · 5.3절 메일 발송 API. 사고 기록 · 인지 왜곡 전용 공개 API는 찾아봤지만 쓸 만한 게 없다(있는 것은 기록 내용을 통째로 보내는 방식이라 5.1절 위반).

- 타이핑 학습 모드로 Go 첫 코드: `go.mod`, `internal/config`(필수 값 모아서 한 번에 거부, 에러에 비밀값 없음), `internal/server`(`GET /api/health`), `cmd/app`(서브커맨드 표 · slog · 우아한 종료). 확인: 테스트 config 4 · server 1 함수 통과, GET 200 / POST 405, Ctrl+C → "서버 종료" 로그.
- 타이핑 학습 모드로 goose · DB 연결: 의존성 pgx · go-redis · goose(사용자 `go get`), `internal/db`(embed 마이그레이션 · pgxpool), `internal/cache`, `app migrate up|down|status`, health가 확인별 ok/error · 503. 확인: migrate pending → applied → down → up, `goose_db_version` 2행(0 · 1), health 200 → Valkey stop 503 → start 200.
- 지금 쓰는 터미널 명령을 README "실행"에 정리했다(사용자 요청). Taskfile이 생기면 같이 고친다.
- 타이핑 학습 모드로 `api/openapi.yaml` 첫 판 + oapi-codegen(`go get -tool`로 go.mod에 고정, strict · std-http). 경로 등록을 생성 코드에 맡기고 `var _ apigen.StrictServerInterface = (*handler)(nil)`로 컴파일 시 확인. 확인: `operationId`를 바꿔 재생성하면 `go build`가 "missing method GetHealthz"로 깨짐 → 되돌리면 통과, server 테스트 4 케이스, curl 200 / 405.
- [설계] `openapi-typescript`는 이번 항목에서 빼고 SvelteKit 항목에서 연결: 아직 `web/`이 없어 설치할 곳이 없다. PLAN 항목에 표시.
- 안내한 `go get github.com/oapi-codegen/runtime`은 필요 없었다: health처럼 파라미터가 없는 경로만 있으면 생성 코드가 runtime을 import하지 않아서 `go mod tidy`가 지웠다. 경로 · 쿼리 파라미터가 생기는 M2 이후에 다시 들어온다.
- 타이핑 학습 모드로 SvelteKit SPA: `sv create`(minimal · TS · prettier · eslint · tailwindcss), adapter-static, shadcn-svelte 1.7(vega · lucide) Button · Card, `layout.css`를 DESIGN 6장 토큰으로, Inter → Pretendard(npm), mode-watcher, openapi-typescript + openapi-fetch, health 화면. 확인: `pnpm check` 0 · `pnpm lint` 통과 · `pnpm build` 성공, 브라우저에서 ok/ok → 라이트 ↔ 다크 → Valkey stop 시 error(아이콘 + 굵은 글자) → Go 종료 시 "연결이 안 돼" → 다시 켜면 ok/ok.
- [설계] 카드 모서리가 DESIGN의 12px이 아니라 14px: shadcn 1.7은 모서리를 곱셈(`md = radius × 0.8`, `xl = × 1.4`)으로 계산한다. `--radius` 10px로 Button(`rounded-md`) 8px을 정확히 맞추고, Card(`rounded-xl`)는 14px로 둔다. 화면을 만들 때 감싸는 부품에서 맞출지 본다(DESIGN 6장 "모양"은 그대로).
- `openapi-typescript` 7.13이 TypeScript `^5`를 피어로 원하는데 프로젝트는 TS 6(`sv` 기본). 생성 도구일 뿐이고 `pnpm gen:api` 결과가 정상이라 그대로 둔다. TS를 낮추면 SvelteKit 3 쪽이 어긋날 위험이 더 크다.
- 타이핑 학습 모드로 `go:embed` + SPA fallback: `internal/webdist`(`all:dist`, `.gitkeep`), `GET /` SPA 핸들러, 오류 봉투 첫 판. 확인: TestSPA 11 케이스, `bin/app.exe serve` 하나로 8080에서 화면 ok/ok · `/checkin`은 서버 200 + SvelteKit 404 화면, `/api/nope` JSON 404, 새로 checkout한 트리(dist에 `.gitkeep`만)에서 `go build` · `go vet` 통과.
- [설계] SPA를 메서드 없는 `/`가 아니라 `GET /`로 등록: `/`로 열면 Go mux가 `POST /api/health`를 경로가 맞는 `/`로 보내 405가 404로 바뀐다. 대가로 없는 `/api/*`에 GET이 아닌 메서드는 405. ARCHITECTURE 1절.
- `.env`는 Go가 읽지 않는다. Taskfile을 만들기 전까지는 PowerShell 한 줄로 터미널에 불러왔고, 이제 Task의 `dotenv`가 읽는다.
- 타이핑 학습 모드로 Taskfile · air: `setup` · `dev`(air + Vite 동시, `output: prefixed`) · `gen` · `test` · `lint` · `build`(빌드 → embed 복사 → `bin/app`) · `serve` · `migrate`. 확인: 환경변수를 지운 터미널에서 `task migrate -- status` 성공, `task build` 뒤 `git status` 깨끗, `task serve` health 200, `task dev` 저장 시 재빌드 · Ctrl+C로 오류 없이 종료. README "실행"을 `task` 중심으로 다시 썼다.
- 타이핑 학습 모드로 golangci-lint(v2.15.0, winget) + GitHub Actions. 첫 실행에서 지적 3개: 테스트의 가짜 비밀번호 G101(테스트 파일의 G101만 제외), 테스트 헬퍼 `httptest.NewRequest`의 noctx(`NewRequestWithContext`로). 확인: `task lint` 0 issues, Actions 첫 실행(run 41001543315) Go · 프론트 · 빌드 세 작업 모두 success — Taskfile의 Linux용 embed 명령과 `.gitkeep` 전략이 깨끗한 Linux에서 처음 검증됐다.
- golangci-lint는 `go get -tool` 대신 winget: 공식 문서가 비권장(검사기 의존성이 `go.mod`와 부딪힐 수 있음). 대신 PC와 `ci.yml`의 버전을 같은 번호로 맞춘다(CLAUDE.md "자주 쓰는 명령").

### 막혔던 일
**`git add --renormalize .` 뒤에도 `.gitattributes`가 Untracked**
- 증상: renormalize를 했는데 `.gitattributes`만 stage되지 않았다.
- 원인: `--renormalize`는 이미 추적 중인 파일만 다시 검사한다. 새 파일은 대상이 아니다.
- 해결: `git add .gitattributes`를 따로. 규칙 자체는 작업 폴더에 파일만 있어도 적용돼 renormalize 결과는 맞았다(바뀐 파일 0개).
- 배운 것: 규칙 파일은 "적용"과 "커밋"이 별개다.

**`psql` 결과 뒤에 `(END)`가 떠서 멈춘 것처럼 보임**
- 원인: `docker compose exec`가 터미널을 붙여 주면 psql이 결과를 페이저(less)로 보여 준다.
- 해결: `q`로 나온다. 한 줄 확인이면 `-tAc` 옵션으로 페이저 없이 값만 받는다.

**`migrate`가 `password authentication failed` (28P01)**
- 증상: `app migrate` · `serve`가 모두 비밀번호 인증 실패. 그런데 `docker compose exec postgres psql`은 잘 된다.
- 원인: `.env`의 `POSTGRES_PASSWORD`는 바꿨는데, 앱 설정 줄을 `.env.example`에서 복사하면서 `DATABASE_URL` 안의 비밀번호는 예시값으로 남았다. 컨테이너 안에서 붙는 psql은 Postgres 이미지가 비밀번호 없이 믿어 줘서(trust) 차이가 안 드러났다.
- 해결: `DATABASE_URL`의 비밀번호를 `POSTGRES_PASSWORD`와 맞추고 `.env`를 다시 불러왔다. 값은 대화에 꺼내지 않고 "같은가"만 비교해서 찾았다.
- 배운 것: 같은 비밀번호가 `.env`에 두 번 들어간다(compose용 · 앱용). README "처음 한 번"에 둘을 맞추라고 적었다. pgx 에러에는 비밀번호가 안 나온다(사용자 · DB 이름만).

**`localhost`가 `[::1]`부터 시도해서 매번 한 번 거절됨**
- 원인: Windows는 `localhost`를 IPv6(`::1`)부터 푸는데 compose는 포트를 `127.0.0.1`에만 열었다.
- 해결: `.env.example`의 `DATABASE_URL` · `VALKEY_URL`을 `127.0.0.1`로.

**안내한 SvelteKit 구조가 실제와 달랐다 (SvelteKit 3)**
- 증상: `svelte.config.js`가 없고, 전역 CSS가 `src/app.css`가 아니라 `src/routes/layout.css`, 별칭이 `$lib`가 아니라 `#lib`.
- 원인: `sv create`가 SvelteKit 3을 설치했다. adapter 설정이 `vite.config.ts`로 옮겨 왔고, 별칭은 Node 표준 subpath imports(`package.json`의 `imports`)로 바뀌었다.
- 해결: 생성 직후 파일을 열어 보고 안내를 고쳤다. shadcn-svelte 1.7 init은 `layout.css` · `#lib`를 스스로 찾았다. import는 확장자까지 적는다(`#lib/...js`).
- 배운 것: 처음 쓰는 도구는 생성기를 돌린 뒤 결과를 먼저 열어 보고 다음 안내를 쓴다. CLAUDE.md "함정"에 적었다.

**`pnpm lint`가 생성된 shadcn Button에서 실패**
- 증상: `ui/button/button.svelte` 63:3 `svelte/no-navigation-without-resolve`.
- 원인: 범용 Button은 `href`를 그대로 쓴다. 링크를 `resolve()`로 만드는 건 쓰는 쪽의 일이다.
- 해결: `eslint.config.js`의 전역 `ignores`에 `ui/` · `schema.d.ts`를 넣었다(`.prettierignore`와 같은 목록). 우리 화면 코드에는 규칙이 그대로 걸린다.

**Tailwind 확장이 있는데도 `layout.css`에 `Unknown at rule` 경고**
- 원인: `sv`가 만든 `web/.vscode/settings.json`(`*.css` → tailwindcss)은 VS Code로 `web` 폴더를 직접 열 때만 읽힌다. 저장소 맨 위를 열어 두면 안 읽힌다.
- 해결: 사용자 설정 또는 저장소 맨 위 `.vscode/settings.json`(커밋 안 됨)에 같은 줄. README "자주 막히는 것"에 적었다.

**`/../../go.mod` 테스트가 200이 아니라 307**
- 원인: Go `ServeMux`는 `..`이 섞인 경로를 핸들러까지 보내지 않고 정리된 주소로 307 리다이렉트한다. 안내한 테스트가 이 동작을 빠뜨리고 200을 기대했다.
- 해결: 307 + `Location: /go.mod`를 확인하는 테스트로 바꿨다. 경로 조작은 mux와 `path.Clean`이 두 겹으로 막는다.

**`task dev`에서 air가 `exclude_dir`의 `internal/webdist/dist`를 무시함**
- 원인: air는 폴더 경로를 운영체제 구분자(Windows는 `\`)로 만들어 `exclude_dir`와 글자 그대로 비교한다(air 소스 `isExcludeDir`). 한 단계 이름(`web` · `tmp`)만 맞았다.
- 해결: `"internal\\webdist\\dist"`(TOML에서 `\`는 `\\`). 같은 때 `build.bin` 지원 중단 경고 → `entrypoint = ["./tmp/app.exe", "serve"]`.

**`task dev`를 Ctrl+C로 끄면 깨진 Y/N 질문 + `exit status 58`**
- 원인 1: `pnpm dev`가 Windows에서 `vite.cmd`(배치 파일)를 거쳐, cmd가 "일괄 작업을 끝내시겠습니까 (Y/N)?"를 물었다(CP949라 Task 출력에서 깨짐).
- 원인 2: 배치를 건너뛰어도 Ctrl+C로 끝난 node의 종료 코드 `0xC000013A`의 마지막 바이트 `0x3A` = 58을 Task가 실패로 읽었다.
- 해결: `dev:web`은 `node node_modules/vite/bin/vite.js dev`로 직접 띄우고 `ignore_error: true`. air 쪽 `Process Exit with Code: 1`(재시작 때 강제 종료)은 Windows에서 정상.

**`webdist.go`에 staticcheck SA9009 경고**
- 원인: 설명 주석이 `// go:embed는 …`으로 시작해서 "공백 때문에 무시된 지시문"으로 보였다(안내한 주석 문장 탓).
- 해결: `// embed 지시문은 …`으로. 다른 `.go` 파일에 `// go:` 같은 주석이 없는지 검색해 진짜 지시문 3개(`//go:embed` 2, 생성 코드 `//go:build` 1)만 있는 걸 확인했다.

**참고: go-redis가 자기 로그를 따로 찍는다**
- Valkey를 끄면 `redis: ... pool.go:770: ... failed to dial` 줄이 slog 형식이 아니라 그대로 나온다. go-redis 내부 로거라서다. 기록 내용은 없어서 지금은 그대로 두고, 운영 로그를 JSON으로 모을 때(M3) 정리할지 본다.

### 이번 세션 커밋
`2d629e5` v1 삭제 · .gitignore / `05e2870` .gitattributes / `276f6af` compose · .env.example / `4494d7a` 문서 갱신 / `2df9652` · `15fcd2f` 외부 API 후보 · A10을 M2로 / `b3df9b8` Go 뼈대 / `5107443` 문서 / `84d4481` goose · health DB 확인 / `6600b4b` README "실행" · 문서 / `33b5e79` openapi · oapi-codegen / `d1d33b2` 문서 / `84646b0` SvelteKit SPA / `73af1fb` 문서 / `93d0e2f` go:embed · SPA fallback / `2517d5b` 문서 / `58b63e4` Taskfile · air / `bb9e918` 주석 SA9009 / `a7bf28b` 문서 / `7b59c92` golangci-lint / `82cafa6` GitHub Actions / (이 커밋) 문서 갱신

### 다음에 할 일
- 집 PC: README "처음 한 번" → `task dev` · `task lint` (M1 완료 기준의 마지막)
- M2 인증: 개념 설명(세션 쿠키 · argon2id · sqlc)부터
- 집 PC: 위 "기기별로 할 일"

---

## 2026-10-29 · 문서 전체 검토 (M0)

### 한 일
- 문서를 다 쓴 뒤 한꺼번에 다시 읽고 서로 대조했다. 불일치 8 · 설계 구멍 6 · 보완 제안 7을 찾아, 보완 1개(`__Host-` 쿠키 접두사)만 빼고 반영했다. 그대로 둔 것: "7칸 / 8단계" 용어, 수정 모드 자동 저장 없음, 계정별 로그인 제한으로 남이 15분 잠글 수 있는 점.

### [설계] 첫 배포를 M6 → M3(인증 바로 뒤)으로
- M4 · M5(옛 M3 · M4)의 완료 기준이 "휴대폰으로 10초 체크인", "휴대폰 → PC 이어 쓰기"인데 배포는 M6이었다. 개발 서버에 휴대폰을 붙이면 `Secure` 쿠키가 `http://IP주소`에서 저장되지 않고, 회사 Wi-Fi는 기기 간 통신을 막는 경우가 많다.
- 인터넷에 여는 건 인증이 있어야 안전하므로 M2 바로 뒤에 빈 앱으로 배포하고, 이후 단계마다 올려서 확인한다. 실제 기록을 쓰기 시작하는 건 그대로 M6. 대안이었던 "개발 중 터널"은 개발 서버를 인터넷에 여는 거라 매번 켜고 끄는 관리가 필요해서 뺐다 (사용자 결정).

### [설계] 검토에서 고친 설계 구멍
- **비밀번호를 바꿔도 다른 세션이 최대 10분 살아 있었다**: DB 세션만 지우고 Valkey 세션 캐시(TTL 10분)는 안 지우는 설계였다. 세션을 지우는 모든 경로(로그아웃 · 비밀번호 변경 · 다른 기기 로그아웃 · 탈퇴)에서 `RETURNING token_hash`로 캐시도 지운다.
- **프록시 뒤 IP**: 배포하면 요청 제한이 모든 사용자를 프록시 IP 하나로 본다 → `TRUSTED_PROXIES` 규칙(ARCHITECTURE 3.3절).
- **로그인 화면이 가입 허용 여부를 알 방법이 없었다** → 공개 설정 `GET /api/config`.
- **D2 "평균 감소율"**: 전 강도 0이면 나눗셈이 안 되고, "감소"는 평가로 읽힌다 → 평균 변화(점). D1 추이는 사고 기록의 "전" 강도만.
- **회사 PC에서 30일 로그인 유지**: 감정 기록을 자리 비운 사이 볼 수 있다 → "로그인 유지" 선택(기본 해제, 해제 시 서버 기한 12시간) + 다른 기기 모두 로그아웃.

### 이번 세션 커밋
(이 커밋) 문서 검토 반영

### 다음에 할 일
- M1 시작: 개념 설명(Go 기초 · Valkey · SvelteKit SPA) → 도구 설치(사용자 직접, 두 PC)

---

## 2026-10-29 · v2 재설계 (M0)

### 한 일
- v1 코드 · 문서를 다시 읽고 문제점 정리 → 전면 재작성 결정 (0001). v1 코드는 태그 `v1-legacy`, 문서는 `archive/`.
- 스택 결정: Go (0002) · SvelteKit SPA + shadcn-svelte (0003) · PostgreSQL (0004) · Valkey (0005), AI 코치 보류 (0006).
- 문서 구조를 새로 잡고 REQUIREMENTS · ARCHITECTURE · DATA_MODEL · DESIGN · PLAN 작성.
- 목업 4장(컬러 시안 · 체크인 · 사고 기록 모바일/데스크톱), 색은 B · 차분한 남색으로 확정.

### [설계] 사용 기기 기준을 "붙잡기는 휴대폰, 정리는 어디서든"으로
- 처음 문서에는 "대부분 휴대폰에서 쓴다"를 확인 없이 가정해서 적었다. 실제로는 회사 PC 앞에 있는 시간이 길다.
- 사고 기록을 성격으로 나눴다: 1 ~ 3단계(상황 · 감정 · 자동 사고)는 떠오른 순간에 붙잡아야 해서 휴대폰, 4 ~ 8단계는 가라앉은 뒤 따져 보는 게 잘 돼서 PC · 휴대폰 둘 다.
- 이미 설계한 서버 쪽 초안 저장(Valkey)으로 기기를 옮겨 이어 쓰기가 그대로 된다 → 새로 만들 것은 [나중에 이어 쓰기] 버튼과 데스크톱 "앞에서 쓴 답" 패널 정도.

### 막혔던 일
**목업이 다크 화면에서 글자가 검게 나옴**
- 증상: 컬러 시안의 다크 절반에서 본문 글자가 배경과 같은 검은색.
- 원인: `color: var(--text)`를 `body`에만 걸었다. CSS 변수는 화면 틀(`.half`)마다 다르게 심었지만, `color`는 `body`에서 **이미 계산된 값**(라이트 색)으로 상속된다.
- 해결: 테마를 심는 틀(`[data-theme]`)에 `color: var(--text)`를 다시 건다.
- 배운 것: 변수를 요소마다 바꿔 끼우는 구조에서는, 그 변수를 쓰는 속성도 그 요소에서 다시 선언해야 한다. 실제 앱도 테마 클래스를 `<html>`에 거는 방식이면 문제없지만, 한 화면에 두 테마를 섞을 땐 같은 함정이 있다.

**`npx http-server`가 설치를 묻는다**
- 증상: 목업을 보려고 `npx http-server`를 쳤더니 "Need to install the following packages"가 나왔다.
- 원인: `npx`는 없는 패키지를 npm 캐시에 임시로 받아 실행한다. 프로젝트 · 전역에 설치되는 것은 아니다.
- 해결: 받아도 되고, conda 환경이면 `python -m http.server 5199`로 받지 않고 띄운다. 목업 README에 둘 다 적음.

### 이번 세션 커밋
`387c9cd` 결정 기록 · REQUIREMENTS / `3c6debe` ARCHITECTURE / `5103e58` DATA_MODEL / `95779fd` DESIGN / `d677587` 사용 기기 기준 / `13ba562` 목업 / `0dc20c5` 색 확정 / `998e3d6` PLAN · DEVLOG / `1122489` CLAUDE.md / (이 커밋) README

### 다음에 할 일
- M1: 개념 설명(Go 기초 · Valkey · SvelteKit SPA) → 도구 설치(사용자 직접)
