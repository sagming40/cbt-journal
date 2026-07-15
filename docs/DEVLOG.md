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