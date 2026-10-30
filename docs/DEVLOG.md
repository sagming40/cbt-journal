# 개발 일지 (v2)

> 겪은 일 · 막힌 일 · 그때 배운 것을 적는다. **최신이 위.** 지난 기록은 고치지 않고 위에 덧붙인다.
> 로드맵 · 진행 상황은 [PLAN.md](PLAN.md), 결정의 이유는 [decisions/](decisions/README.md). v1 일지는 [archive/DEVLOG_v1.md](archive/DEVLOG_v1.md).

## 현재 상태
| 항목 | 내용 |
|---|---|
| 단계 | **M3 첫 배포 진행 중** — 서버 `cbt-1` · 기본 보안 · Docker 끝, 다음은 운영 구성 설계 (PLAN 2장) |
| 끝난 것 | M0 · M1 · M2 전부(두 PC에서 확인). M2 인증: `users` · `sessions`, 세션 쿠키 로그인(로그인 유지 · 쿠키 재발급 · Valkey 캐시), 요청 제한 4종, 가입(플래그) · 내 정보 · 비밀번호 변경 · 다른 기기 로그아웃, 유출 비밀번호 거절, 기본 잠금(openapi security), `task user` 관리 명령, 프론트 로그인 · 가입 · 설정 · 401 처리 · 메뉴, Vitest, CI 통합 테스트 · 생성 코드 확인 |
| 진행 중 | M3: **운영 구성 설계안을 사용자에게 보여 줌 — 결정 3개 대기** (아래 2026-10-30 "운영 구성 설계안" 블록) |
| 다음에 할 일 | 결정 3개 받기 → 설계 문서(ARCHITECTURE 7절 · SERVER.md · PLAN) 먼저 → 저장소 운영 파일(`deploy/`) · `task build:linux` · `task deploy` · `task backup:pull` (Claude 자동) → 서버 작업(사용자) → DNS → 첫 배포 → 완료 기준 확인 |
| 기기별로 할 일 | **회사 PC**: 서버 접속용 SSH 키를 따로 만들어 서버 `authorized_keys`에 추가 — 순서는 [SERVER.md](SERVER.md) 3.4절 (Vultr에 등록만 해서는 이미 만든 서버에 안 들어간다) |
| **기억할 날짜** | **2026-11-30 Vultr 크레딧 $250 만료** — 그 전에 테스트 서버 삭제(Destroy, 정지는 요금이 계속 나감), 12월부터 서버 한 대 값이 카드로. **같은 날 Vultr 자동 백업($1/월) 유지 여부 판단**(우리 DB 백업이 믿을 만하면 끔). 2027-10-30 `sagming.com` 자동 갱신 (Cloudflare). 2029-04 Ubuntu 24.04 지원 종료 → 새 서버로 옮기기 |

## 작성 규칙
- 세션(작업 단위)이 끝날 때 날짜 블록을 맨 위에 추가하고, 위 "현재 상태"를 고친다.
- 막혔던 일은 해결됐어도 남긴다. 형식: **증상 / 원인 / 해결 / 배운 것**.
- 구현 중에 설계를 바꿨으면 `[설계]` 항목으로 이유를 적고, 해당 설계 문서의 변경 이력도 고친다.
- 블록 끝에 **이번 세션 커밋**과 **다음에 할 일**을 꼭 쓴다.

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
