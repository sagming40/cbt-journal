// server/src/controllers/authController.js

const bcrypt = require('bcrypt');
const pool = require('../config/db');
const jwt = require('jsonwebtoken'); // 로그인 API

// 회원가입 처리 함수
exports.register = async (req, res) => {
  const { email, password, nickname } = req.body;  

  // 1. 필수값 체크 (텅 빈 값으로 요청오면 바로 컷)
  if (!email || !password || !nickname) {
    return res.status(400).json({
      success: false,
      message: '이메일, 비밀번호, 닉네임을 모두 입력해주세요.',  
    });
  }

  try {
    // 2. 이메일 중복 확인
    //    비유: 창고(DB)에 가서 "이 이메일 이미 등록된 사람있어?" 먼저 물어보는 것
    const existingUser = await pool.query(
      'SELECT id FROM users WHERE email = $1',
      [email]  
    );

    if (existingUser.rows.length > 0) {
      return res.status(400).json({
        success: false,
        message: '이미 존재하는 이메일입니다.',
      });  
    }

    // 3. 비밀번호 암호화
    //    비유: 원본 비밀번호를 "지문 도장"으로 찍어버리는 것. 이후엔 원본으로 되돌릴 수 없음.
    const saltRounds = 10; // 도장을 얼마나 복잡하게 찍을지 (숫자가 높을수록 안전하지만 느려짐)
    const hashedPassword = await bcrypt.hash(password, saltRounds);
    
    // 4. DB에 새 회원 저장
    await pool.query(
      'INSERT INTO users (email, password, nickname) VALUES ($1, $2, $3)',
      [email, hashedPassword, nickname]  
    );

    // 5. 성공 응답
    return res.status(201).json({
      success: true,
      message: '회원가입이 완료되었습니다.',  
    });
  } catch (err) {
    console.error(err);
    return res.status(500).json({
      success: false,
      message: '서버 오류가 발생했습니다.',  
    });
  }
};

// 로그인 처리 함수
exports.login = async (req, res) => {
  const { email, password } = req.body;

  if (!email || !password) {
    return res.status(400).json({
      success: false,
      message: '이메일과 비밀번호를 입력해주세요.',  
    });
  }

  try {
    // 1. 이메일로 회원 찾기
    const result = await pool.query(
      'SELECT id, email, password, nickname FROM users WHERE email = $1',
      [email]  
    );

    const user = result.rows[0];

    // 2. 회원이 아예 없는 경우
    if (!user) {
      return res.status(401).json({
        success: false,
        message: '이메일 또는 비밀번호가 일치하지 않습니다.',
      });  
    }

    // 3. 비밀번호 대조
    //    비유: 입력한 비번을 똑같은 방식으로 다시 도장 찍어서, 저장된 지문이랑 모양이 같은지 비교하는 것
    const isPasswordMatch = await bcrypt.compare(password, user.password);

    if (!isPasswordMatch) {
      return res.status(401).json({
        success: false,
        message: '이메일 또는 비밀번호가 일치하지 않습니다.',
      });  
    }

    // 4. JWT 팔찌 발급
    //    비유: 신분증(user.id) 정보를 봉인해서 위조 방지 도장까지 찍은 손목 팔찌를 만들어주는 것
    const token = jwt.sign(
      { id: user.id, email: user.email }, // 팔찌 안에 담을 정보 (payload)
      process.env.JWT_SECRET,             // 워조 방지 도장 (비밀 키)
      { expiresIn: '24h' }                // 팔찌 유효기간      
    );

    // 5. 성공 응답 (토큰 + 유저 정보 반환)
    return res.status(200).json({
      success: true,
      token,
      user: {
        id: user.id,
        email: user.email,
        nickname: user.nickname,
      },  
    });
  } catch (err) {
    console.error(err);
    return res.status(500).json({
      success: false,
      message: '서버 오류가 발생했습니다.'  
    });
  }
};

// 현재 로그인한 유저 정보 조회 (미들웨어 테스트용)
exports.getMe = async (req, res) => {
  try {
    // req.user는 authMiddleware를 통과했을 때만 존재함 (authMiddleware에서 심어준 것)
    const result = await pool.query(
      'SELECT id, email, nickname, created_at FROM users WHERE id = $1',
      [req.user.id]  
    );

    return res.status(200).json({
      success: true,
      user: result.rows[0],  
    });
  } catch (err) {
    console.error(err);
    return res.status(500).json({
      success: false,
      message: '서버 오류가 발생했습니다.'  
    });
  }  
};

