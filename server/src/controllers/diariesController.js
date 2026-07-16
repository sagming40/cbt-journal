// server/src/controllers/diariesController.js

const pool = require('../config/db');

// CBT 일기 저장하기
exports.createDiary = async (req, res) => {
  const {
    pre_emotion_type,
    pre_intensity,
    situation,
    automatic_thought,
    alternative_thought,
    post_intensity,
    distortion_ids // 예: [1, 3]
  } = req.body;  

  const userId = req.user.id; // authMiddleware가 심어준 로그인 사용자 id

  const client = await pool.connect();

  try {
    await client.query('BEGIN');

    // 1. diaries 테이블에 본문 저장
    const diaryResult = await client.query(
      `INSERT INTO diaries
        (user_id, pre_emotion_type, pre_intensity, situation, automatic_thought, alternative_thought, post_intensity)
       VALUES ($1, $2, $3, $4, $5, $6, $7)
       RETURNING id`,
      [
        userId,
        pre_emotion_type,
        pre_intensity,
        situation,
        automatic_thought,
        alternative_thought,
        post_intensity,
      ]   
    );

    const diaryId = diaryResult.rows[0].id;

    // 2. 선택한 왜곡 개수만큼 diary_distortions에 매핑 저장
    if (Array.isArray(distortion_ids) && distortion_ids.length > 0) {
      for (const distortionId of distortion_ids) {
        await client.query(
          `INSERT INTO diary_distortions (diary_id, distortion_id)
           VALUES ($1, $2)`,
          [diaryId, distortionId]   
        );
      }  
    }

    await client.query('COMMIT');

    res.status(201).json({
      success: true,
      message: '오늘의 일기가 성공적으로 기록되었습니다.',
      diary_id: diaryId,  
    });
  } catch (err) {
    await client.query('ROLLBACK');
    console.error(err);
    res.status(500).json({
      success: false,
      message: '일기 저장 중 오류가 발생했습니다.',  
    });
  } finally {
    client.release();
  }
};

// 일기 전체 목록 조회 (GET /api/diaries)
exports.getDiaryList = async (req, res) => {
  try {
    const userId = req.user.id; // 미들웨어가 심어둔 검증 정보

    const result = await pool.query(
      `SELECT id, created_at, pre_emotion_type, pre_intensity, post_intensity, automatic_thought
       FROM diaries
       WHERE user_id = $1
       ORDER BY created_at DESC`,
      [userId] 
    );

    res.status(200).json(result.rows);
  } catch (err) {
    console.error(err);
    res.status(500).json({ success: false, message: '서버 오류가 발생했습니다.' });
  }
};

// 일기 상세 조회 (GET /api/diaries/:id)
exports.getDiaryDetail = async (req, res) => {
  try {
    const userId = req.user.id;
    const diaryId = req.params.id;

    const result = await pool.query(
      `SELECT d.id, d.user_id, d.pre_emotion_type, d.pre_intensity, d.situation,
              d.automatic_thought, d.alternative_thought, d.post_intensity, d.created_at,
              array_agg(cd.name) FILTER (WHERE cd.name IS NOT NULL) AS distortions
       FROM diaries d
       LEFT JOIN diary_distortions dd ON dd.diary_id = d.id
       LEFT JOIN cognitive_distortions cd ON cd.id = dd.distortion_id
       WHERE d.id = $1
       GROUP BY d.id`,
      [diaryId] 
    );

    // 아예 존재하지 않는 id
    if (result.rows.length === 0) {
      return res.status(404).json({ success: false, message: '해당 일기를 찾을 수 없습니다.' });
    }

    const diary = result.rows[0];

    // 존재는 하지만 남의 일기인 경우 (보안: 404로 통일해서 "있는지 없는지"도 숨김)
    if (diary.user_id !== userId) {
      return res.status(404).json({ success: false, message: '해당 일기를 찾을 수 없습니다' });
    }

    delete diary.user_id; // 응답엔 굳이 보여주지 않아도 되는 필드라 제거
    res.status(200).json(diary);
  } catch (err) {
    console.error(err);
    res.status(500).json({ success: false, message: '서버 오류가 발생했습니다.' });
  }
};
