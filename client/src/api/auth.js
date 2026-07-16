// client/src/api/auth.js

import instance from "./axios";

// 회원가입 요청
export const registerUser = async (email, password, nickname) => {
  const response = await instance.post('/auth/register', {
    email,
    password,
    nickname,
  });
  
  return response.data;
};

// 로그인 요청
export const loginUser = async (email, password) => {
  const response = await instance.post('/auth/login', {
    email,
    password,
  });
  
  return response.data;
}