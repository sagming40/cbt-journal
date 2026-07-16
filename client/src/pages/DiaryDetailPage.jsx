// client/src/pages/DiaryDetailPage.jsx

import { useState, useEffect } from "react";
import { useParams, Link } from "react-router-dom";
import { getDiaryDetail } from "../api/diaries";

function DiaryDetailPage() {
  const { id } = useParams();
  
  const [diary, setDiary] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchDiary = async () => {
      try {
        const data = await getDiaryDetail(id);
        setDiary(data);
      } catch (err) {
        console.error(err);
        setError('일기를 불러오는 데 실패했습니다.');
      } finally {
        setLoading(false);
      }  
    };

    fetchDiary();
  }, [id]);

  if (loading) {
    return <p>불러오는 중...</p>;
  }

  if (error) {
    return <p>{error}</p>;
  }

  return (
    <div>
      <Link to="/diaries">← 목록으로</Link>

      <h1>일기 상세 보기</h1>
      <p>{new Date(diary.created_at).toLocaleDateString()}</p>

      <h3>1. 사전 감정</h3>
      <p>{diary.pre_emotion_type} (강도: {diary.pre_intensity})</p>

      <h3>2. 상황</h3>
      <p>{diary.situation}</p>

      <h3>3. 자동 사고</h3>
      <p>{diary.automatic_thought}</p>

      <h3>4. 인지 왜곡</h3>
      <p>{diary.distortions.join(', ')}</p>

      <h3>5. 대안 사고</h3>
      <p>{diary.alternative_thought}</p>

      <h3>6. 사후 감정</h3>
      <p>강도: {diary.post_intensity}</p>  
    </div>
  );
}

export default DiaryDetailPage;