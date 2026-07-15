// server/src/routes/diariesRoutes.js

const express = require('express');
const router = express.Router();
const { createDiary } = require('../controllers/diariesController');
const authMiddleware = require('../middlewares/authMiddleware');

router.post('/', authMiddleware, createDiary);

module.exports = router;
