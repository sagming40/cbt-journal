const pool = require('../config/db');

// 인지 왜곡 유형 전체 목록 가져오기
exports.getDistortions = async (req, res) => {
  try {
    const result = await pool.query(
      'SELECT id, name FROM cognitive_distortions ORDER BY id ASC'  
    );
    res.status(200).json(result.rows);
  } catch (err) {
    console.error(err);
    res.status(500).json({
      success: false,
      message: '왜곡 목록을 불러오는 중 오류가 발생했습니다.',  
    });
  }  
};
