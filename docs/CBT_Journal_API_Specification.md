# 🔌 CBT 기반 마음 챙김 감정 기록 저널 - 4단계: API 명세서 (API Specification)

본 문서는 React(프론트엔드)와 Node.js(백엔드) 간에 데이터를 원활하게 주고받기 위해 작성된 **4단계: API 명세서**입니다. 모든 API의 요청과 응답은 JSON 형식을 기준으로 하며, 인증이 필요한 API는 HTTP Header에 JWT 토큰을 실어서 보내야 합니다.

*※ 로그인 토큰이 필요한 모든 API는 HTTP Header에 `Authorization: Bearer <JWT_TOKEN>`을 필수로 포함해야 합니다.*

---

## 1. 회원 관리 API (Authentication)

### ① 이메일 회원가입 (Register)
* **메서드 및 URL**: `POST /api/auth/register`
* **설명**: 새로운 사용자를 서비스에 등록합니다.
* **요청 데이터 (Request Body)**:
  ```json
  {
    "email": "user@test.com",
    "password": "password123!",
    "nickname": "코딩초보"
  }
  ```
* **응답 데이터 (Response)**:
  * **성공 (201 Created)**:
    ```json
    {
      "success": true,
      "message": "회원가입이 완료되었습니다."
    }
    ```
  * **실패 (400 Bad Request - 이메일 중복 등)**:
    ```json
    {
      "success": false,
      "message": "이미 존재하는 이메일입니다."
    }
    ```

### ② 이메일 로그인 (Login)
* **메서드 및 URL**: `POST /api/auth/login`
* **설명**: 이메일과 비밀번호를 검증하고, 향후 API 요청 시 사용할 JWT 인증 토큰을 발급받습니다.
* **요청 데이터 (Request Body)**:
  ```json
  {
    "email": "user@test.com",
    "password": "password123!"
  }
  ```
* **응답 데이터 (Response)**:
  * **성공 (200 OK)**:
    ```json
    {
      "success": true,
      "token": "eyJhbGciOiJIUzI1NiIsIn...",
      "user": {
        "id": 1,
        "email": "user@test.com",
        "nickname": "코딩초보"
      }
    }
    ```
  * **실패 (401 Unauthorized)**:
    ```json
    {
      "success": false,
      "message": "이메일 또는 비밀번호가 일치하지 않습니다."
    }
    ```

---

## 2. 인지 왜곡 마스터 API (Distortions)

### ① 인지 왜곡 유형 목록 가져오기 (Get Distortion List) 🆕
* **메서드 및 URL**: `GET /api/distortions`
* **설명**: 일기 작성 화면에서 인지 왜곡 체크박스를 그리기 위해, `cognitive_distortions` 테이블에 미리 정의된 전체 왜곡 유형 목록을 가져옵니다.
* **호출 시점**: 일기 작성 화면 진입 시 (또는 앱 최초 로딩 시 캐싱)
* **인증**: 상수 데이터이므로 토큰이 없어도 되나, 일관성을 위해 로그인 후 호출을 권장합니다.
* **응답 데이터 (Response)**:
  * **성공 (200 OK)**:
    ```json
    [
      { "id": 1, "name": "흑백논리" },
      { "id": 2, "name": "과도한 일반화" },
      { "id": 3, "name": "파국화" },
      { "id": 4, "name": "정신적 필터" },
      { "id": 5, "name": "감정적 추론" }
    ]
    ```

> 💡 **왜 이 API가 필요한가?**
> 일기 작성 시 서버로는 왜곡의 **이름**이 아니라 **번호**(`distortion_ids: [1, 3]`)를 보냅니다. 따라서 프론트엔드는 "흑백논리 = 1번, 파국화 = 3번"이라는 매핑 정보를 서버로부터 먼저 받아와야 체크박스를 올바른 id와 함께 렌더링할 수 있습니다.

---

## 3. CBT 감정 일기 API (Diaries)

### ① 일기 작성하기 (Create Diary)
* **메서드 및 URL**: `POST /api/diaries`
* **설명**: 사용자가 작성한 5단계 CBT 기록 데이터를 데이터베이스에 저장합니다.
* **요청 데이터 (Request Body)**:
  ```json
  {
    "pre_emotion_type": "불안",
    "pre_intensity": 8,
    "situation": "실습 과제를 제출했는데 피드백이 없음",
    "automatic_thought": "과제를 못 해서 나를 무시하는 걸까?",
    "alternative_thought": "단지 교수님이 바쁘셔서 확인을 못 하신 걸 거야.",
    "post_intensity": 3,
    "distortion_ids": [1, 3]
  }
  ```
  * `distortion_ids`: 선택한 인지 왜곡 고유 ID 리스트 (예: 1=흑백논리, 3=파국화). `GET /api/distortions`로 받아온 id를 사용합니다.
* **응답 데이터 (Response)**:
  * **성공 (201 Created)**:
    ```json
    {
      "success": true,
      "message": "오늘의 일기가 성공적으로 기록되었습니다.",
      "diary_id": 42
    }
    ```
  * **실패 (401 Unauthorized - 토큰 누락/만료)**:
    ```json
    {
      "success": false,
      "message": "인증에 실패했습니다. 다시 로그인해 주세요."
    }
    ```

### ② 일기 전체 목록 가져오기 (Get Diary List)
* **메서드 및 URL**: `GET /api/diaries`
* **설명**: 현재 로그인한 사용자가 작성한 일기 목록 전체를 최신 작성일 순으로 가져옵니다. (메인 화면 렌더링용)
* **응답 데이터 (Response)**:
  * **성공 (200 OK)**:
    ```json
    [
      {
        "id": 42,
        "created_at": "2026-07-15T03:00:00.000Z",
        "pre_emotion_type": "불안",
        "pre_intensity": 8,
        "post_intensity": 3,
        "automatic_thought": "과제를 못 해서 나를 무시하는 걸까?..."
      }
    ]
    ```
  * **실패 (401 Unauthorized - 토큰 누락/만료)**:
    ```json
    {
      "success": false,
      "message": "인증에 실패했습니다. 다시 로그인해 주세요."
    }
    ```

### ③ 일기 상세 조회하기 (Get Diary Detail)
* **메서드 및 URL**: `GET /api/diaries/:id`  *(예: `/api/diaries/42`)*
* **설명**: 특정 일기 하나의 세부 작성 내용을 모두 가져옵니다. (상세 보기 화면용)
* **응답 데이터 (Response)**:
  * **성공 (200 OK)**:
    ```json
    {
      "id": 42,
      "pre_emotion_type": "불안",
      "pre_intensity": 8,
      "situation": "실습 과제를 제출했는데 피드백이 없음",
      "automatic_thought": "과제를 못 해서 나를 무시하는 걸까?",
      "alternative_thought": "단지 교수님이 바쁘셔서 확인을 못 하신 걸 거야.",
      "post_intensity": 3,
      "created_at": "2026-07-15T03:00:00.000Z",
      "distortions": ["흑백논리", "파국화"]
    }
    ```
    * `distortions`: DB JOIN을 거쳐 문자열 배열로 반환합니다.
  * **실패 (404 Not Found - 존재하지 않는 ID)**:
    ```json
    {
      "success": false,
      "message": "해당 일기를 찾을 수 없습니다."
    }
    ```

### ④ 일기 삭제하기 (Delete Diary) 🆕
* **메서드 및 URL**: `DELETE /api/diaries/:id`  *(예: `/api/diaries/42`)*
* **설명**: 로그인한 사용자가 자신이 작성한 특정 일기를 삭제합니다. (연결된 `diary_distortions` 데이터는 ON DELETE CASCADE로 함께 삭제됨)
* **보안 주의**: 삭제 전 반드시 해당 일기의 `user_id`가 요청한 사용자의 토큰 id와 일치하는지 확인해야 합니다. (남의 일기 삭제 방지)
* **응답 데이터 (Response)**:
  * **성공 (200 OK)**:
    ```json
    {
      "success": true,
      "message": "일기가 삭제되었습니다."
    }
    ```
  * **실패 (403 Forbidden - 본인 일기가 아님)**:
    ```json
    {
      "success": false,
      "message": "삭제 권한이 없습니다."
    }
    ```
  * **실패 (404 Not Found - 존재하지 않는 ID)**:
    ```json
    {
      "success": false,
      "message": "해당 일기를 찾을 수 없습니다."
    }
    ```

> ℹ️ **참고**: 일기 삭제(④)는 기능 정의서의 필수(Must-Have)에는 없던 기능이나, 실제 서비스 완성도를 위해 추가 제안한 API입니다. MVP 범위를 좁게 가져가려면 나중 단계로 미뤄도 됩니다.

---

## 4. 대시보드 통계 API (Dashboard)

### ① 감정 통계 데이터 가져오기 (Get Dashboard Stats)
* **메서드 및 URL**: `GET /api/dashboard/stats`
* **설명**: 대시보드의 요약 위젯, 꺾은선 그래프, 도넛 차트에 채워질 가공 데이터를 조회합니다.
* **응답 데이터 (Response)**:
  * **성공 (200 OK)**:
    ```json
    {
      "summary": {
        "total_count": 12,
        "average_reduction": 45.2
      },
      "weekly_trends": [
        { "date": "07-10", "pre": 8, "post": 3 },
        { "date": "07-12", "pre": 6, "post": 2 }
      ],
      "distortions_ratio": [
        { "name": "파국화", "count": 15 },
        { "name": "흑백논리", "count": 10 },
        { "name": "정신적 필터", "count": 5 }
      ]
    }
    ```
    * `average_reduction` 공식: `(사전 강도 평균 - 사후 강도 평균) / 사전 강도 평균 * 100`
  * **실패 (401 Unauthorized - 토큰 누락/만료)**:
    ```json
    {
      "success": false,
      "message": "인증에 실패했습니다. 다시 로그인해 주세요."
    }
    ```

---

## 📌 API 요약표

| 분류 | 메서드 | URL | 인증 | 비고 |
| :--- | :--- | :--- | :--- | :--- |
| 회원 | POST | `/api/auth/register` | ❌ | 회원가입 |
| 회원 | POST | `/api/auth/login` | ❌ | 로그인 (토큰 발급) |
| 왜곡 | GET | `/api/distortions` | △ | 왜곡 유형 목록 🆕 |
| 일기 | POST | `/api/diaries` | ✅ | 일기 작성 |
| 일기 | GET | `/api/diaries` | ✅ | 일기 목록 |
| 일기 | GET | `/api/diaries/:id` | ✅ | 일기 상세 |
| 일기 | DELETE | `/api/diaries/:id` | ✅ | 일기 삭제 🆕 |
| 통계 | GET | `/api/dashboard/stats` | ✅ | 대시보드 통계 |
