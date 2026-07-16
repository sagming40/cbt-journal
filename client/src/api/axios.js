// client/src/api/axios.js

import axios from 'axios';

// 1. axios 기본 설정을 가진 인스턴스 생성
//    비유: 매번 새 우체국 창구를 찾아가는 대신,
//    "우리 서버 전용 창구"를 하나 고정으로 만들어두는 것
const instance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL, // 예: http://localhost:4000/api  
});

// 2. 요청 인터셉터 (Request Interceptor)
//    비유: 편지를 보내기 직전에 거치는 검문소.
//    "너 신분증(토큰) 있어? 있으면 편지 봉투에 붙여서 보내줄게"
instance.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');

    if (token) {
      config.headers.Authorization = `Bearer ${token}`;  
    }

    return config;
  },
  (error) => {
    return Promise.reject(error);
  }  
);

export default instance;