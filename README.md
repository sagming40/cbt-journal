# 🧠 마음 챙김 감정 기록 저널 (CBT Thought Record)

> 인지행동치료(CBT) 이론을 기반으로, 사용자가 스스로의 **부정적 자동 사고**를 인지하고
> **합리적인 대안 사고**로 감정을 조절하도록 돕는 감정 기록 저널 웹 애플리케이션입니다.

![React](https://img.shields.io/badge/React-20232A?style=flat&logo=react&logoColor=61DAFB)
![Node.js](https://img.shields.io/badge/Node.js-339933?style=flat&logo=node.js&logoColor=white)
![Express](https://img.shields.io/badge/Express-000000?style=flat&logo=express&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-4169E1?style=flat&logo=postgresql&logoColor=white)
![JWT](https://img.shields.io/badge/JWT-000000?style=flat&logo=jsonwebtokens&logoColor=white)

---

## 📌 프로젝트 소개

우리는 나쁜 일이 생겼을 때 순간적으로 왜곡된 생각(자동 사고)을 떠올리고, 그 생각 때문에 감정이 더 나빠지곤 합니다.
이 서비스는 인지행동치료의 **사고 기록법(Thought Record)** 5단계를 디지털화하여, 사용자가 자신의 생각 패턴을 스스로 점검하고 감정을 다스릴 수 있도록 돕습니다.

### 핵심 가치
1. **감정 변화의 정량적 측정** — 기록 전(사전)/후(사후) 감정 강도를 수치로 남겨 변화를 눈으로 확인합니다.
2. **왜곡된 생각의 인지와 반박 훈련** — 자동 사고에 담긴 인지 왜곡 유형을 짚고, 이성적인 대안 사고를 작성합니다.
3. **통계를 통한 메타인지** — 누적 데이터를 차트로 시각화하여 자신의 생각 패턴을 한눈에 파악합니다.

---

## ✨ 주요 기능

| 구분 | 기능 | 설명 |
| :--- | :--- | :--- |
| **회원 관리** | 회원가입 / 로그인 / 로그아웃 | 이메일 기반 인증, 비밀번호 `bcrypt` 암호화, `JWT` 토큰으로 로그인 상태 유지 |
| **CBT 일기 작성** | 5단계 사고 기록 | 사전 감정 강도 → 상황·자동 사고 → 인지 왜곡 유형 선택 → 대안 사고 → 사후 감정 강도 |
| **일기 조회** | 목록 / 상세 보기 | 작성일 최신순 목록, 카드 클릭 시 전체 내용 상세 조회 |
| **감정 통계 대시보드** | 추이 그래프 / 왜곡 비율 차트 | 사전·사후 감정 변화 추이(선 그래프), 자주 겪는 인지 왜곡 비율(도넛 차트) |
| **일기 삭제** | 본인 일기 삭제 | 소유자 검증 후 삭제 (연결된 왜곡 데이터는 `CASCADE`로 함께 삭제) |

---

## 🛠️ 기술 스택

| 영역 | 기술 |
| :--- | :--- |
| **Frontend** | React, React Router, Axios, Recharts (차트) |
| **Backend** | Node.js, Express |
| **Database** | PostgreSQL |
| **인증** | JWT (JSON Web Token), bcrypt |
| **협업/설계** | Figma (와이어프레임), dbdiagram.io (ERD) |

---

## 🗄️ 데이터베이스 구조 (ERD)

4개의 테이블로 구성되며, `diaries`와 `cognitive_distortions`는 `diary_distortions` 브릿지 테이블을 통해 **N:M(다대다)** 관계를 맺습니다.

```
users (회원)
  └─< diaries (일기)
        └─< diary_distortions (일기-왜곡 연결) >─┐
                                                cognitive_distortions (인지 왜곡 마스터)
```

- **users** — 회원 계정 및 인증 정보
- **diaries** — CBT 일기 본문 및 사전/사후 감정 수치
- **cognitive_distortions** — 인지 왜곡 유형 마스터(상수) 데이터
- **diary_distortions** — 일기와 왜곡을 잇는 매핑 테이블 (`UNIQUE(diary_id, distortion_id)`로 중복 방지)

> 📄 상세 스키마는 [`/docs/CBT_Journal_ERD.md`](./docs/CBT_Journal_ERD.md) 참고

---

## 🔌 API 요약

| 분류 | 메서드 | URL | 인증 | 설명 |
| :--- | :--- | :--- | :---: | :--- |
| 회원 | `POST` | `/api/auth/register` | ❌ | 회원가입 |
| 회원 | `POST` | `/api/auth/login` | ❌ | 로그인 (토큰 발급) |
| 왜곡 | `GET` | `/api/distortions` | △ | 인지 왜곡 유형 목록 |
| 일기 | `POST` | `/api/diaries` | ✅ | 일기 작성 |
| 일기 | `GET` | `/api/diaries` | ✅ | 일기 목록 |
| 일기 | `GET` | `/api/diaries/:id` | ✅ | 일기 상세 |
| 일기 | `DELETE` | `/api/diaries/:id` | ✅ | 일기 삭제 |
| 통계 | `GET` | `/api/dashboard/stats` | ✅ | 대시보드 통계 |

> 📄 요청/응답 상세 명세는 [`/docs/CBT_Journal_API_Specification.md`](./docs/CBT_Journal_API_Specification.md) 참고

---

## 📁 폴더 구조

```
cbt-journal/
├── client/                 # React (프론트엔드)
│   ├── src/
│   │   ├── api/            # 서버 통신 (axios 인스턴스 + 요청 함수)
│   │   ├── components/     # 재사용 UI 컴포넌트
│   │   ├── context/        # 전역 상태 (로그인 토큰 등)
│   │   ├── pages/          # 화면 단위 (로그인, 작성, 목록, 상세, 대시보드)
│   │   └── App.jsx         # 라우팅 및 인증 가드
│   ├── .env               # 환경 변수 (git 제외)
│   └── package.json
│
├── server/                 # Node.js + Express (백엔드)
│   ├── src/
│   │   ├── config/        # DB 연결 설정
│   │   ├── controllers/   # 요청 처리 로직
│   │   ├── middlewares/   # JWT 인증 미들웨어 등
│   │   ├── routes/        # API 라우팅
│   │   └── app.js
│   ├── db/                # 테이블 생성 SQL (init.sql)
│   ├── .env               # 환경 변수 (git 제외)
│   └── package.json
│
├── docs/                   # 설계 문서 (기능 정의서, ERD, API 명세서, 개발 일지)
└── README.md
```

---

## 🚀 실행 방법

### 사전 준비
- Node.js 18 이상
- PostgreSQL 14 이상

### 1. 저장소 클론
```bash
git clone https://github.com/사용자명/cbt-journal.git
cd cbt-journal
```

### 2. 데이터베이스 세팅
```bash
# PostgreSQL에서 데이터베이스 생성
createdb cbt_journal

# 테이블 생성 및 초기 데이터(인지 왜곡 유형) 삽입
# (docs의 ERD를 기반으로 작성한 SQL 스크립트 실행 예정)
```

### 3. 백엔드 실행
```bash
cd server
npm install

# .env 파일 생성 후 아래 값 설정
# PORT=4000
# DATABASE_URL=postgresql://user:password@localhost:5432/cbt_journal
# JWT_SECRET=your_secret_key

npm run dev
```

### 4. 프론트엔드 실행
```bash
cd client
npm install

# .env 파일 생성 후 아래 값 설정
# VITE_API_BASE_URL=http://localhost:4000/api

npm run dev
```

---

## 🗺️ 개발 로드맵

> 기준일: 2026.07.16

| 단계 | 내용 | 백엔드 | 프론트엔드 | 비고 |
| :--- | :--- | :---: | :---: | :--- |
| **Phase 0** | 환경 세팅 (DB · 서버 · 프론트 뼈대) | ✅ | ✅ | 프론트 뼈대(Vite)는 07.16 완료 |
| **Phase 1** | 회원 관리 (가입 · 로그인 · JWT 인증) | ✅ | ⬜ | 로그인 / 회원가입 페이지 미착수 |
| **Phase 2** | 일기 작성 (왜곡 목록 조회 + 저장) | ✅ | ⬜ | 일기 작성 페이지 미착수 |
| **Phase 3** | 일기 조회 (목록 · 상세 보기) | ✅ | ✅ | 07.16 완료 |
| **Phase 4** | 대시보드 (통계 집계 + 차트) | ⬜ | ⬜ | |
| **Phase 5** | 삭제 · 에러 처리 · UI 마무리 | ⬜ | ⬜ | |
| **Phase 6** | 확장 기능 (소셜 로그인, 검색, 캘린더 등) | ⬜ | ⬜ | 여유되면 확장 |

**원칙** — 각 단계는 `[DB → 백엔드 → 프론트엔드]`를 관통하는 하나의 완성된 기능 단위로 개발합니다.

**현재 상태** — Phase 1 · 2는 백엔드만 선행되어 위 원칙에서 이탈한 상태이며, 프론트엔드 회수 작업을 진행 중입니다.
Phase 3 프론트엔드(목록 · 상세)를 먼저 구현했기 때문에, 다음 순서는 Phase 1 프론트엔드(로그인 / 회원가입)입니다.

> DB 4개 테이블(`users` / `diaries` / `cognitive_distortions` / `diary_distortions`)은 Phase 0에서 일괄 생성되어
> 이후 스키마 변경이 없으므로 별도 열로 표기하지 않았습니다.

---

## 👤 개발자

- **사공민규** — 한국폴리텍II대학 인천캠퍼스 컴퓨터공학과 (하이테크과정)
- 개발 기간: 2026년 7월 ~ (진행 중)

---

## 📄 관련 문서

- [기능 정의서](./docs/CBT_Journal_Feature_List.md)
- [ERD 설계서](./docs/CBT_Journal_ERD.md)
- [API 명세서](./docs/CBT_Journal_API_Specification.md)
- [개발 일지](./docs/DEVLOG.md)
