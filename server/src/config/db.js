// server/src/config/db.js

// 1. .env 파일에 적어둔 비밀 메모(DATABASE_URL, JWT_SECRET 등)를 읽어오는 도구를 불러옴
require('dotenv').config();

// 2. pg 패키지에서 "커넥션 풀(Pool)"이라는 걸 꺼내옴
const { Pool } = require('pg');

// 3. 커넥션 풀 생성
//    비유: 트럭을 한 대만 두는 게 아니라, 여러 대를 미리 대기시켜두는 '차고지'를 만들기
//    요청이 여러 개 동시에 들어와도 트럭(연결)을 매번 새로 안 만들고 대기 중인 요청을 바로 씀 -> 훨씬 빠름
const pool = new Pool({
  connectionString: process.env.DATABASE_URL,  
});

// 4. 다른 파일에서 이 pool을 갖다 쓸 수 있게 내보냄
module.exports = pool;
