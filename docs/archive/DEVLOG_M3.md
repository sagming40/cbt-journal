# 개발 일지 — M3 · 첫 배포

> **M3 기간의 DEVLOG 날짜 블록**을 옮겨 둔 문서다. 보관만 하고 고치지 않는다(링크 경로만 이 폴더에 맞췄다).
> 지금 일지는 [../DEVLOG.md](../DEVLOG.md), 단계와 완료 기준은 [../PLAN.md](../PLAN.md).

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
- 서버 명령을 정리한 런북 [SERVER.md](../SERVER.md) 새로 만듦(`78d1f26`).

### 다음에 할 일
- 위 결정 3개 → 진행 순서 1부터. 회사 PC에서 이어 하면 먼저 [SERVER.md](../SERVER.md) 3.4절로 회사 PC 열쇠를 서버에 추가(집 PC로 접속해 둔 상태가 아니면, 회사에서 Vultr 웹 콘솔 비상구로 넣거나 집에서 미리 추가).

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
