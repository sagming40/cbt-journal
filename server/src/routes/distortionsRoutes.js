const express = require('express');
const router = express.Router();
const { getDistortions } = require('../controllers/distortionsController');

router.get('/', getDistortions);

module.exports = router;
