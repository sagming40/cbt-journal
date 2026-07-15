-- ============================================================
-- CBT 저널 프로젝트 - 초기 테이블 생성 스크립트
-- 기준 문서: docs/CBT_Journal_ERD.md (v1.1)
-- 실행 순서: users -> cognitive_distortions -> diaries -> diary_distortions
-- ============================================================

-- 1. users (회원 테이블)
CREATE TABLE users (
    id            SERIAL PRIMARY KEY,
    email         VARCHAR(100) UNIQUE NOT NULL,
    password      VARCHAR(255) NOT NULL,
    nickname      VARCHAR(50) NOT NULL,
    created_at    TIMESTAMP DEFAULT NOW()
);

-- 2. cognitive_distortions (인지 왜곡 마스터 테이블 - 상수 데이터)
CREATE TABLE cognitive_distortions (
    id      SERIAL PRIMARY KEY,
    name    VARCHAR(50) UNIQUE NOT NULL
);

-- 3. diaries (일기 테이블)
CREATE TABLE diaries (
    id                    SERIAL PRIMARY KEY,
    user_id               INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pre_emotion_type      VARCHAR(50) NOT NULL,
    pre_intensity         INT NOT NULL,
    situation             TEXT NOT NULL,
    automatic_thought     TEXT NOT NULL,
    alternative_thought   TEXT NOT NULL,
    post_intensity        INT NOT NULL,
    created_at            TIMESTAMP DEFAULT NOW()
);

-- 성능 팁: user_id로 "이 회원이 쓴 일기 전부" 조회가 잦으므로 인덱스 생성
CREATE INDEX idx_diaries_user_id ON diaries(user_id);

-- 4. diary_distortions (일기-왜곡 연결 테이블, N:M 브릿지)
CREATE TABLE diary_distortions (
    id             SERIAL PRIMARY KEY,
    diary_id       INT NOT NULL REFERENCES diaries(id) ON DELETE CASCADE,
    distortion_id  INT NOT NULL REFERENCES cognitive_distortions(id) ON DELETE RESTRICT,
    UNIQUE (diary_id, distortion_id)
);

-- ============================================================
-- 초기 데이터(Seed): 인지 왜곡 유형 10종
-- 서비스 시작 시 딱 한 번만 넣어두는 상수 데이터
-- ============================================================
INSERT INTO cognitive_distortions (name) VALUES
    ('흑백논리'),
    ('과도한 일반화'),
    ('정신적 필터'),
    ('긍정적인 것 무시하기'),
    ('성급한 결론(독심술/예언자적 오류)'),
    ('확대와 축소'),
    ('감정적 추론'),
    ('당위적 진술'),
    ('낙인찍기'),
    ('개인화'),
    ('파국화');
