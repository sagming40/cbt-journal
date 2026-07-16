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

## 2026-07-16 · Phase 3 일기 조회 (목록 · 상세 보기) — 백엔드

### 오늘 한 일
- 일기 목록 조회 API (`GET /api/diaries`) — 로그인한 사용자의 일기를 최신순으로 조회
- 일기 상세 조회 API (`GET /api/diaries/:id`) — 일기 본문 + `diary_distortions`/`cognitive_distortions` JOIN으로 왜곡 이름 배열(`distortions`) 함께 반환
- 남의 일기 조회 시도 시 403이 아닌 404로 통일 응답 (존재 여부 자체를 숨겨서 보안 강화)
- Postman으로 4가지 케이스 테스트 완료: 목록 조회 / 상세 조회 / 존재하지 않는 id / 타인 일기 접근 차단

### 겪었던 이슈들

**1. `diaryController is not defined`**
라우터 파일에서 컨트롤러를 불러올 때 `const { createDiary } = require(...)`처럼 중괄호로 특정 함수 하나만 꺼내놓고, 정작 사용할 땐 `diaryController.getDiaryList`처럼 객체 전체인 것처럼 접근하려 해서 발생한 에러. 중괄호(`{ }`)는 "모듈에서 특정 값만 골라 꺼내기"이고, 중괄호 없이 받으면 "모듈 전체를 객체로 받기"라는 차이를 이번에 확실히 짚고 넘어갔다. `const diaryController = require('../controllers/diariesController');`로 변경해서 해결.

**2. `Cannot read properties of undefined (reading 'id')`**
남의 일기 접근을 막는 권한 체크 로직에서 `if (diary.user.id !== userId)`라고 작성했는데, SQL에서 뽑아온 컬럼명은 `user_id`(언더스코어)라서 `diary` 객체엔 애초에 `user`라는 속성이 없었다. `diary.user`가 `undefined`인 상태에서 그 뒤에 `.id`를 또 읽으려다 보니 에러가 남. `diary.user_id !== userId`로 수정. DB 컬럼명(`snake_case`)을 JS에서 그대로 쓸 때는 점(`.`)이 아니라 언더스코어(`_`)로 이어져 있다는 걸 헷갈리지 않도록 다시 새겼다.

**3. Postman 테스트 중 id 번호 혼동**
남의 일기 접근 테스트를 할 때, 새로 만든 테스트 계정이 실제로 작성한 일기 id를 착각해서(`2`번인 줄 알았는데 실제로는 `3`번) 엉뚱한 id로 요청을 보낸 적이 있었다. 결과적으로 404가 뜨긴 했지만, 그게 "없는 id라서" 뜬 건지 "남의 일기라서" 막힌 건지 구분이 안 되는 상황이었음. `psql`로 `SELECT id, user_id FROM diaries ORDER BY id;`를 직접 돌려서 각 일기가 어느 회원 소유인지 확인한 뒤에야 정확히 의도한 시나리오(타인 접근 차단)를 검증할 수 있었다. 테스트 결과를 신뢰하려면 "무엇을 검증하고 있는지"부터 정확히 짚고 가야 한다는 걸 다시 느낌.

### 오늘 배운 것 / 느낀 점
- `require`할 때 중괄호 유무 차이(구조 분해 할당 vs 객체 전체 참조)를 명확히 구분하게 됨.
- SQL 컬럼명(`snake_case`)과 JS 객체 접근 문법을 섞어 쓸 때 오타(`.` vs `_`)에 특히 취약하다는 걸 체감. 앞으로 DB 관련 필드를 다룰 땐 컬럼명을 한 번 더 확인하는 습관을 들이기로 함.
- 보안 테스트(타인 데이터 접근 차단)를 검증할 때는 "어떤 조건에서 어떤 응답이 나와야 하는지"를 미리 명확히 정의해두지 않으면, 결과가 우연히 맞아떨어져도 진짜 검증이 안 된 상태일 수 있다는 걸 배움. `psql`로 직접 데이터를 확인하는 습관이 중요하다는 것도 느낌.

### 다음에 할 일
- Phase 3 프론트엔드: 일기 목록 페이지 + 상세 보기 페이지 구현

---

## 2026-07-16 · Phase 3 일기 조회 (목록 · 상세 보기) — 프론트엔드

### 오늘 한 일
- `client` 폴더에 Vite로 React 프로젝트 스캐폴딩 (린터는 ESLint 선택), `react-router-dom` / `axios` 설치
- `src` 하위 폴더 구조 정리 — `api` / `components` / `context` / `pages`
- `.env`에 `VITE_API_BASE_URL` 설정 후 axios 인스턴스(`api/axios.js`) 작성 — 요청 인터셉터로 JWT 토큰 자동 첨부
- 서버 통신 함수 분리 — `api/auth.js`(회원가입/로그인), `api/diaries.js`(목록/상세 조회)
- `context/AuthContext.jsx` — `AuthProvider`로 토큰·유저 정보를 전역 관리, `useAuth` 커스텀 훅 제공
- `App.jsx` — `BrowserRouter` 라우팅 설정 + `PrivateRoute`로 비로그인 접근 차단
- `pages/DiaryListPage.jsx` — 목록 조회 후 카드 렌더링, `<Link>`로 상세 페이지 연결
- `pages/DiaryDetailPage.jsx` — `useParams`로 URL의 `id`를 꺼내 상세 조회, `distortions` 배열은 `join(', ')`으로 출력
- 로그인 페이지가 아직 없어, 브라우저 콘솔에서 `localStorage`에 토큰을 직접 주입해 목록 → 상세 이동까지 동작 확인 완료

### 겪었던 이슈들

**1. 프로젝트 원칙 이탈을 뒤늦게 발견**
`client` 폴더를 열어보니 완전히 비어 있었다. README에는 "각 Phase는 `[DB → 백엔드 → 프론트엔드]`를 관통하는 완성된 기능 단위로 개발한다"고 적어놨는데, 실제로는 Phase 1 · 2 · 3의 백엔드만 연달아 만들고 프론트엔드는 하나도 손대지 않은 상태였다. 로드맵에는 Phase 1 · 2가 ✅로 표시돼 있어서 "다 끝난 줄" 알고 있었지만, 사실은 반쪽만 완료된 것이었다. **문서와 실제 코드가 어긋나면 문서는 오히려 나를 속인다**는 걸 체감했고, 로드맵 표를 백엔드/프론트엔드 열로 분리해 다시 표기했다.

**2. `localStorage.setItem('token', '')` — 토큰을 넣었는데도 로그인 페이지로 튕김**
임시 테스트를 위해 콘솔에서 토큰을 넣었는데, 값이 빈 문자열(`''`)인 채로 실행됐다. `PrivateRoute`의 `if (!token)` 검사에서 빈 문자열은 **falsy**로 취급되기 때문에, 키는 존재하지만 "토큰 없음"과 완전히 동일하게 동작했다. JS에서 `''`, `0`, `null`, `undefined`, `NaN`이 전부 falsy라는 걸 실제 버그로 마주치니 확실히 각인됐다.

**3. 크롬 DevTools 콘솔에 붙여넣기가 막힘**
토큰을 붙여넣으려 하니 "Don't paste code into the DevTools Console that you don't understand..." 경고가 뜨면서 입력 자체가 되지 않았다. 최근 크롬에 추가된 보안 기능으로, 콘솔에 코드를 붙여넣게 유도하는 사기(self-XSS)를 막기 위한 장치라고 한다. 콘솔에 `allow pasting`을 직접 타이핑해서 해제한 뒤 정상 진행.

### 오늘 배운 것 / 느낀 점
- **axios 인터셉터의 가치**를 체감했다. `axios.js`에서 토큰 첨부를 한 번 처리해두니, `api/diaries.js`의 조회 함수들이나 각 페이지에서는 토큰을 전혀 신경 쓰지 않아도 됐다. 만약 페이지마다 헤더를 붙였다면 나중에 토큰 관리 방식을 바꿀 때 수정할 곳이 흩어졌을 것이다.
- **`useEffect`의 의존성 배열**이 `[]`와 `[id]`로 갈리는 이유를 이해했다. 목록 페이지는 처음 한 번만 불러오면 되지만(`[]`), 상세 페이지는 URL의 `id`가 바뀌면 다시 불러와야 하므로 `[id]`가 필요하다. 이걸 빠뜨리면 "링크는 눌렀는데 화면이 안 바뀌는" 버그가 생긴다.
- **로딩 상태(`loading`)를 따로 두는 이유**도 알게 됐다. 서버 응답이 오기 전 찰나에 빈 배열이 렌더링되면서 "아직 작성한 일기가 없어요"라는 잘못된 메시지가 순간 보일 수 있는데, `loading` 상태로 그 구간을 가려줘야 한다.
- 무엇보다, DB → API → 화면까지 하나의 데이터가 한 줄기로 이어지는 걸 눈으로 처음 확인했다. 그동안 Postman의 JSON으로만 보던 데이터가 실제 화면에 카드로 뜨는 순간, 백엔드와 프론트엔드가 어떻게 맞물리는지가 확실히 잡혔다.

### 다음에 할 일
- Phase 1 프론트엔드: 로그인 / 회원가입 페이지 구현 (지금은 `App.jsx`에 임시 `<div>`로만 자리를 잡아둔 상태)
- 이후 Phase 2 프론트엔드(일기 작성 페이지)로 이어서 원칙 이탈 회수 마무리

---
