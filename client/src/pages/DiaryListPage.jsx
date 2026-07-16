// client/src/pages/DiaryListPage.jsx

import { useState, useEffect } from "react";
import { Link } from 'react-router-dom';
import { getDiaries } from "../api/diaries";

function DiaryListPage() {
  const [diaries, setDiaries] = useState([]);   // 서버에서 받아온 일기 목록 
  const [loading, setLoading] = useState(true); // 데이터 로딩 중
  const [error, setError] = useState(null);     // 에러
  
  // 페이지가 처음 화면에 나타날 때 딱 한 번, 서버에 목록을 요청한다.
  useEffect(() => {
    const fetchDiaries = async () => {
      try {                  // 먼저 try
        const data = await getDiaries();
        setDiaries(data);    // 성공 ㅡ 일기 목록 출력
      } catch (err) {        // 실패하면 catch 처리
        console.error(err);  // 실패 ㅡ 에러 메세지 출력
        setError('일기 목록을 불러오는 데 실패했습니다.');
      } finally {            // 성공/실패 여부 상관 없이 무조건 실행
        setLoading(false);   // 일기 목록/에러 메세지 출력으로 어느 쪽이든 끝이 나면 ㅡ 로딩 끝남 표시
      }
    };

    fetchDiaries();
  }, []);

  // 로딩 중 ㅡ 빈 목록 대신, 지금 데이터를 가져오는 중임을 사용자에게 암시
  // 이 과정이 없으면 순간적으로 잘못된 메세지가 출력될 수 있음.
  if (loading) {
    return <p>불러오는 중...</p>
  }

  if (error) {
    return <p>{error}</p>;
  }

  return (
    <div>
      <h1>내 감정 기록 목록</h1>

      {diaries.length === 0 ? (
        <p>아직 작성한 일기가 없어요.</p>
      ) : (
        diaries.map((diary) => (  // 배열(diary)에 들어 있는 일기 데이터를 하나씩 카드로 변환
            
          // key={diary.id} ㅡ 몇 번 id의 일기인지 표시해두는 이름표
          // ㅡ 목록이 바뀔 때(추가/삭제) 이 카드가 몇 번 id의 카드인지 구분 하는데 ※필요※
          // <Link key={...} to={`/diaries/${diary.id}`}>
          //  React Router 전용 링크 ㅡ 새로고침 없이 화면만 매끄럽게 전환
          // XX <a> 태그 > 클릭할 때마다 브라우저가 페이지를 통째로 새로고침(느림, 상태도 다 날아감) XX   
          <Link key={diary.id} to={`/diaries/${diary.id}`}>
            <div style={{ border: '1px solid #ccc', padding: '12px', marginBottom: '8px' }}>   
              <p>{new Date(diary.created_at).toLocaleDateString()}</p> 
              <p>
                {diary.pre_emotion_type} {diary.pre_intensity} → {diary.post_intensity}
              </p>
              <p>{diary.automatic_thought}</p>  
            </div>
          </Link>  
        ))
      )} 
    </div>
  );
}

export default DiaryListPage;
