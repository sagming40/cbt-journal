// server/src/routes/diariesRoutes.js

const express = require('express');
const router = express.Router();
const diaryController = require('../controllers/diariesController');
const authMiddleware = require('../middlewares/authMiddleware');

router.post('/', authMiddleware, diaryController.createDiary);
router.get('/', authMiddleware, diaryController.getDiaryList);
router.get('/:id', authMiddleware, diaryController.getDiaryDetail);

module.exports = router;
