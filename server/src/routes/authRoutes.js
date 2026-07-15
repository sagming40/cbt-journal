// server/src/routes/authRoutes.js

const express = require('express');
const router = express.Router();
const authController = require('../controllers/authController');
const authMiddleware = require('../middlewares/authMiddleware'); // Middleware 검증

// POST /api/auth/register
router.post('/register', authController.register);  // 회원가입
router.post('/login', authController.login);        // 로그인
router.get('/me', authMiddleware, authController.getMe); // 유저 정보 조회 (Middleware 검증)

module.exports = router;
