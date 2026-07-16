// client/src/api/diaries.js

import instance from "./axios";

// 일기 목록 전체 가져오기
export const getDiaries = async () => {
  const response = await instance.get('/diaries');
  
  return response.data;
};

// 일기 상세 하나 가져오기
export const getDiaryDetail = async (id) => {
  const response = await instance.get(`/diaries/${id}`);
  
  return response.data;
};