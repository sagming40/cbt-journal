# 🗄️ CBT 기반 마음 챙김 감정 기록 저널 - 3단계: ERD 설계서 (PostgreSQL)

본 문서는 인지행동치료(CBT) 기반 감정 기록 저널의 데이터베이스 테이블 구조를 정의한 **3단계: ERD 설계서**입니다. 
관계형 데이터베이스(RDBMS)인 PostgreSQL을 기준으로 작성되었습니다.

---

## 1. 데이터베이스 테이블 요약 관계도
* **Users (회원)** `1 : N` **Diaries (일기)**: 한 명의 회원은 여러 개의 일기를 작성할 수 있습니다.
* **Diaries (일기)** `1 : N` **Diary_Distortions (일기-왜곡 연결)**: 하나의 일기는 여러 개의 인지 왜곡 유형과 매핑될 수 있습니다.
* **Cognitive_Distortions (인지 왜곡 마스터)** `1 : N` **Diary_Distortions (일기-왜곡 연결)**: 하나의 인지 왜곡 유형은 여러 일기에서 참조될 수 있습니다.
* *즉, Diaries와 Cognitive_Distortions는 Diary_Distortions를 매개로 하는 **N:M (다대다) 관계**를 형성합니다.*

---

## 2. 테이블 상세 정의서

### ① `users` (회원 테이블)
사용자의 계정 정보와 로그인 인증 데이터를 보관합니다.

| 컬럼명 | 데이터 타입 | 제약 조건 | 설명 |
| :--- | :--- | :--- | :--- |
| **id** | SERIAL (INT) | Primary Key | 고유 회원 번호 (자동 증가) |
| **email** | VARCHAR(100) | UNIQUE, NOT NULL | 로그인용 이메일 (중복 불가) |
| **password** | VARCHAR(255) | NOT NULL | 암호화된 비밀번호 |
| **nickname** | VARCHAR(50) | NOT NULL | 서비스에서 사용할 닉네임 |
| **created_at** | TIMESTAMP | DEFAULT NOW() | 계정 생성 일시 |

### ② `diaries` (일기 테이블)
사용자가 작성하는 CBT 일기 본문 및 감정 수치 데이터를 저장합니다.

| 컬럼명 | 데이터 타입 | 제약 조건 | 설명 |
| :--- | :--- | :--- | :--- |
| **id** | SERIAL (INT) | Primary Key | 고유 일기 번호 (자동 증가) |
| **user_id** | INT | Foreign Key (users.id), **NOT NULL** | 일기를 작성한 회원 번호 |
| **pre_emotion_type** | VARCHAR(50) | NOT NULL | 사전 감정 종류 (불안, 슬픔 등) |
| **pre_intensity** | INT | NOT NULL | 사전 감정 강도 (1 ~ 10) |
| **situation** | TEXT | NOT NULL | 발생한 상황 |
| **automatic_thought**| TEXT | NOT NULL | 부정적 생각 (자동 사고) |
| **alternative_thought**| TEXT | NOT NULL | 이성적 반박 (대안 사고) |
| **post_intensity** | INT | NOT NULL | 사후 감정 강도 (1 ~ 10) |
| **created_at** | TIMESTAMP | DEFAULT NOW() | 일기 작성 일시 |

> 💡 **성능 팁**: `user_id`는 "이 회원이 쓴 일기 전부 가져와" 조회에 항상 쓰이는 컬럼이라, 인덱스(index)를 걸어두면 데이터가 쌓여도 조회가 빨라집니다. (예: `CREATE INDEX idx_diaries_user_id ON diaries(user_id);`)

### ③ `cognitive_distortions` (인지 왜곡 마스터 테이블)
CBT에서 정의하는 10여 가지 인지 왜곡 종류를 미리 정의해 두는 상수 테이블입니다.

| 컬럼명 | 데이터 타입 | 제약 조건 | 설명 |
| :--- | :--- | :--- | :--- |
| **id** | SERIAL (INT) | Primary Key | 왜곡 고유 번호 (자동 증가) |
| **name** | VARCHAR(50) | UNIQUE, NOT NULL | 왜곡 명칭 (예: 흑백논리, 파국화, 정신적 필터) |

> 💡 **초기 데이터(Seed)**: 이 테이블은 사용자가 채우는 게 아니라, 서비스 시작 시점에 개발자가 왜곡 유형 10여 개를 미리 `INSERT` 해두는 "상수 테이블"입니다. 프론트엔드는 이 목록을 `GET /api/distortions`로 불러와 체크박스로 그립니다.

### ④ `diary_distortions` (일기-왜곡 연결 테이블)
하나의 일기에 여러 개의 인지 왜곡이 선택될 때 발생하는 **다대다(N:M) 관계**를 정규화하여 풀기 위한 브릿지 테이블입니다.

| 컬럼명 | 데이터 타입 | 제약 조건 | 설명 |
| :--- | :--- | :--- | :--- |
| **id** | SERIAL (INT) | Primary Key | 연결 데이터 고유 ID (자동 증가) |
| **diary_id** | INT | Foreign Key (diaries.id), NOT NULL | 해당 일기 번호 |
| **distortion_id** | INT | Foreign Key (cognitive_distortions.id), NOT NULL | 연결된 인지 왜곡 번호 |
| — | — | **UNIQUE (diary_id, distortion_id)** | **같은 일기에 같은 왜곡이 중복 저장되는 것을 방지하는 복합 유니크 제약** |

> ⚠️ **중요**: `UNIQUE (diary_id, distortion_id)` 복합 제약이 없으면, 버그나 중복 요청 시 `일기42-파국화`가 여러 줄 쌓일 수 있고, 이 경우 대시보드의 왜곡 통계 개수가 실제보다 부풀려집니다. DB 차원에서 "한 일기에 한 왜곡은 한 번만"을 강제합니다.

---

## 3. 삭제 정책 (ON DELETE)
데이터의 정합성을 위해 부모 데이터가 삭제될 때의 동작을 아래와 같이 정의합니다.

| 관계 | 삭제 정책 | 이유 |
| :--- | :--- | :--- |
| `diaries.user_id → users.id` | **CASCADE** | 회원 탈퇴 시, 그 회원이 작성한 일기도 함께 삭제됩니다. (주인 없는 일기 방지) |
| `diary_distortions.diary_id → diaries.id` | **CASCADE** | 일기 삭제 시, 그 일기에 연결된 왜곡 매핑도 함께 삭제됩니다. |
| `diary_distortions.distortion_id → cognitive_distortions.id` | **RESTRICT** | 왜곡 마스터는 상수 데이터이므로, 일기가 참조 중이면 함부로 삭제되지 않도록 막습니다. |

---

## 4. DBML (dbdiagram.io 복사용 코드)
아래의 코드를 [dbdiagram.io](https://dbdiagram.io/)에 붙여넣으면 시각화된 ERD 다이어그램을 즉시 확인할 수 있습니다.

```dbml
// 1. 회원 테이블
Table users {
  id integer [primary key, increment]
  email varchar(100) [unique, not null]
  password varchar(255) [not null]
  nickname varchar(50) [not null]
  created_at timestamp [default: `now()`]
}

// 2. 일기 테이블
Table diaries {
  id integer [primary key, increment]
  user_id integer [not null]
  pre_emotion_type varchar(50) [not null]
  pre_intensity integer [not null]
  situation text [not null]
  automatic_thought text [not null]
  alternative_thought text [not null]
  post_intensity integer [not null]
  created_at timestamp [default: `now()`]

  Indexes {
    user_id [name: 'idx_diaries_user_id']
  }
}

// 3. 인지 왜곡 마스터 테이블
Table cognitive_distortions {
  id integer [primary key, increment]
  name varchar(50) [unique, not null]
}

// 4. 다대다 관계 해결을 위한 매핑 테이블
Table diary_distortions {
  id integer [primary key, increment]
  diary_id integer [not null]
  distortion_id integer [not null]

  Indexes {
    (diary_id, distortion_id) [unique] // 같은 일기 + 같은 왜곡 중복 저장 방지
  }
}

// 5. 관계 정의 (삭제 시 동작 포함)
Ref: diaries.user_id > users.id [delete: cascade]
Ref: diary_distortions.diary_id > diaries.id [delete: cascade]
Ref: diary_distortions.distortion_id > cognitive_distortions.id [delete: restrict]
```
