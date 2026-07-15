# 📓 개발 일지 (Dev Log)

CBT 저널 프로젝트를 진행하며 그날그날 배운 것, 겪은 문제, 고민했던 것들을 기록합니다.

---

## 2026-07-15 · Phase 1 회원 관리 (회원가입 / 로그인 / JWT)

### 오늘 한 일
- `server` 폴더 초기 세팅 (express, pg, bcrypt, jsonwebtoken, dotenv, cors 설치)
- `.env` + DB 연결 모듈(`config/db.js`) 구성, `/api/health`로 연결 테스트
- 회원가입 API (`POST /api/auth/register`) — bcrypt로 비밀번호 해싱 후 저장
- 로그인 API (`POST /api/auth/login`) — 비밀번호 대조 후 JWT 발급
- 인증 미들웨어(`authMiddleware.js`) + 보호된 라우트(`GET /api/auth/me`) 구현
- Postman으로 전체 API 성공/실패 케이스 테스트 완료

### 겪었던 이슈들

**1. `npm run dev`가 자꾸 실패함 (`ENOENT: no such file or directory`)**
`package.json`이 있는 `server` 폴더가 아니라 다른 위치(루트 혹은 `server/src`)에서 명령어를 실행해서 벌어진 일이었다. npm 스크립트는 항상 `package.json`이 있는 폴더 기준으로 실행해야 한다는 걸 확실히 배웠다.

**2. Postman에서 계속 `401 Unauthorized`**
토큰 값은 맞게 넣었는데 `Bearer ` 접두사를 빼먹었었다. `Authorization: Bearer <token>` 형식은 그냥 관례가 아니라, 서버 코드에서 `authHeader.startsWith('Bearer ')`로 직접 체크하고 있었기 때문에 접두사가 없으면 검증 단계까지 가지도 못하고 바로 걸러진다는 걸 알게 됐다.

**3. `/api/auth/me`에서 `500 Internal Server Error`**
`TypeError: Cannot read properties of undefined (reading 'id')`. 알고 보니 `getMe` 함수에서 `req.user.id`라고 써야 할 걸 `res.user.id`로 오타 냈던 것 (r → s 한 글자 차이). `req`(들어오는 요청)와 `res`(나가는 응답)가 이름이 비슷해서 헷갈렸는데, 미들웨어에서 심어둔 커스텀 데이터는 항상 **요청 쪽(`req`)**에 있다는 걸 확실히 기억하게 됐다.

### 오늘 배운 것 / 느낀 점
- 에러 메시지를 그대로 읽고, 스택트레이스에서 정확히 몇 번째 줄인지 짚어내는 흐름이 생각보다 중요하다는 걸 체감했다. 무작정 코드를 다시 보는 것보다 에러 로그부터 정확히 읽는 습관을 들여야겠다.
- `req` / `res` 처럼 이름이 비슷한 것들은 실수하기 쉬우니, 앞으로 코드 짤 때 "이게 들어오는 것(req)인지 나가는 것(res)인지" 한 번 더 생각하고 타이핑하는 습관을 들이자.
- JWT 인증 흐름(발급 → 헤더에 실어 보냄 → 미들웨어에서 검증 → 컨트롤러로 전달)을 직접 처음부터 끝까지 만들어보니, 그동안 막연했던 "로그인 유지가 어떻게 되는 거지?"라는 개념이 확실히 잡혔다.

### 다음에 할 일
- Phase 2: 일기 작성 (인지 왜곡 목록 조회 + CBT 일기 저장 API)

---

## 2026-07-15 · Phase 2 일기 작성 (왜곡 목록 조회 + CBT 일기 저장)

### 오늘 한 일
- 인지 왜곡 유형 조회 API (`GET /api/distortions`) — cognitive_distortions 테이블 전체를 id 순으로 조회
- CBT 일기 저장 API (`POST /api/diaries`) — diaries 본문 INSERT + diary_distortions 매핑 INSERT를 트랜잭션으로 묶어 처리
- nodemon 설치 및 dev 스크립트 적용 (`node` → `nodemon`)
- pgAdmin으로 diaries, diary_distortions 두 테이블에 데이터가 동시에 저장되는 것을 눈으로 확인

### 겪었던 이슈들

**1. `GET /api/distortions` 호출 시 `Cannot GET /api/distortions` (404)**
컨트롤러, 라우터, app.js 등록까지 코드 자체는 처음부터 문제가 없었다. 원인은 `package.json`의 `dev` 스크립트가 `nodemon`이 아니라 그냥 `node src/app.js`였던 것. nodemon 없이 `node`로만 서버를 띄우면 파일을 아무리 고쳐도 서버가 그 사실을 전혀 모른다는 걸 알게 됐다. 코드를 고칠 때마다 서버를 수동으로 껐다 켜야 했던 거였고, 그래서 새로 만든 라우트가 반영이 안 됐던 것. `nodemon` 설치 후 `dev` 스크립트를 `nodemon src/app.js`로 바꿔서 해결했다.

**2. `git add .`로 한 번에 커밋해버려서 커밋 단위가 뒤섞임**
"인지 왜곡 API 추가"랑 "nodemon 설정 변경"을 따로 커밋하려 했는데, `git add .`로 전체 파일을 한 번에 스테이징한 뒤 `git commit`을 두 번 치니 첫 번째 커밋에서 5개 파일이 전부 들어가버렸다. `push`하기 전이라 `git reset --soft HEAD~1`로 커밋만 되돌리고, `git restore --staged .`로 스테이징을 풀어서 파일 단위로 다시 나눠 커밋했다. 앞으로는 커밋을 나눠야 할 때 `git add .` 대신 파일을 직접 지정해서 스테이징하는 습관을 들이기로 했다.

**3. 트랜잭션(Transaction) 처음 적용**
일기 저장 API는 `diaries`와 `diary_distortions` 두 테이블을 동시에 다뤄야 해서, 중간에 실패하면 데이터가 반쯤만 저장되는 문제가 생길 수 있었다. `BEGIN` → 여러 INSERT → `COMMIT`(성공 시) / `ROLLBACK`(실패 시) 구조로 묶어서 처리했고, 이때 `pool.query()`가 아니라 `pool.connect()`로 얻은 `client`를 계속 재사용해야 같은 트랜잭션으로 묶인다는 것도 배웠다. `client.release()`를 `finally` 블록에 넣어서 항상 커넥션을 반납하도록 한 것도 이번에 처음 신경 쓴 부분.

### 오늘 배운 것 / 느낀 점
- `nodemon` 없이 개발하면 "코드 고쳐도 반영 안 되는" 상황이 반복될 수밖에 없다는 걸 직접 겪고 나서야 왜 다들 nodemon을 기본으로 까는지 이해가 됐다.
- Git에서 "스테이징(add)"과 "커밋(commit)"이 분리되어 있는 이유를 이번에 체감했다 — `add`를 세밀하게 하지 않으면 아무리 커밋 메시지를 나눠 써도 소용이 없다는 걸 알게 됐다.
- 트랜잭션은 "여러 단계가 하나의 논리적 작업으로 묶여야 할 때" 쓰는 거라는 걸 diaries/diary_distortions 케이스로 직접 만들어보고 나서야 확실히 이해했다. 계좌이체처럼 "다 되거나, 하나도 안 되거나" 둘 중 하나여야 하는 상황에서 필수라는 게 이제 명확해졌다.

### 다음에 할 일
- Phase 3: 일기 조회 (목록 · 상세 보기)

---