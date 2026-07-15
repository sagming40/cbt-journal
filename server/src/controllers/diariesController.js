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
