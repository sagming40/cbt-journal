// server/src/app.js

const express = require('express');
const cors = require('cors');
const pool = require('./config/db');
const authRoutes = require('./routes/authRoutes'); // ⭐ route 연결

const app = express();

app.use(cors());          // React(다른 포트)에서 오는 요청 허용
app.use(express.json());  // 요청 body를 JSON으로 자동 해석
app.use('/api/auth', authRoutes); // ⭐ route 연결

// 연결 테스트용 라우트
app.get('/api/health', async (req, res) => {
  try {
    const result = await pool.query('SELECT NOW()');
    res.json({ success: true, dbTime: result.rows[0].now });
  } catch (err) {
    console.error(err);
    res.status(500).json({ success: false, message: 'DB 연결 실패' });
  }  
});

const PORT = process.env.PORT || 4000;
app.listen(PORT, () => {
  console.log(`🚀 서버 실행 중: http://localhost:${PORT}`);  
});
