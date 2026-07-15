// server/src/middlewares/authMiddleware.js

const jwt = require('jsonwebtoken');

// 비유: 놀이공원 입구에 서 있는 검표원
//       "팔찌 보여주세요" -> 팔찌 있으면 통과, 없거나 가짜면 입장 거부
module.exports = function authMiddleware(req, res, next) {
  // 1. 요청 헤더에서 토큰 꺼내기
  //    형식: "Authorization: Bearer eyJhbGci..."
  const authHeader = req.headers['authorization'];
  
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return res.status(401).json({
      success: false,
      message: '인증에 실패했습니다. 다시 로그인 해주세요.',  
    });
  }

  // 2. "Bearer " 뒷부분만 잘라내기 (진짜 토큰 값)
  const token = authHeader.split(' ')[1];

  try {
    // 3. 토큰이 진짜 우리 서버가 발급한 게 맞는지, 만료는 안 됐는지 검증
    const decoded = jwt.verify(token, process.env.JWT_SECRET);
   
    // 4. 검증 통과하면, 그 안에 담겨있던 유저 정보를 req.user에 실어서
    //    다음 단계(실제 컨트롤러)로 넘겨줌
    req.user = decoded; // { id, email, iat, exp }

    next(); // "통과! 다음으로 가세요"
  } catch (err) {
    // 서명이 틀렸거나, 만료됐거나, 아예 이상한 문자열이면 여기로 떨어짐
    return res.status(401).json({
      success: false,
      message: '인증에 실패했습니다. 다시 로그인 해주세요.',  
    });
  } 
};