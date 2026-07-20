// client/src/pages/Login.jsx

import { useState } from "react";
import { useNavigate } from 'react-router-dom';
import { loginUser } from '../api/auth';
import { useAuth } from '../context/AuthContext';
import './login.css'

function Login() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [errorMsg, setErrorMsg] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (e) => {
    e.preventDefault();
    setErrorMsg('');
    setIsLoading(true);

    try {
      // auth.js의 loginUser는 email, password를 따로따로 받는 함수라서
      // 객체로 묶지 않고 그대로 두 개를 넘겨줌
      const data = await loginUser(email, password);
      login(data.token, data.user); // AuthContext에 토큰 + 유저정보 저장
      navigate('/'); // 로그인 성공하면 메인으로 이동    
    } catch (err) {
      const message = err.response?.data?.message || '로그인 중 오류가 발생했어요.';
      setErrorMsg(message);  
    } finally {
      setIsLoading(false); // 성공/실패 여부 상관없이 로딩 표시등 off  
    }
  };

  return (
    <div className="login-page">
      <h2>로그인</h2>
      <form onSubmit={handleSubmit}>
        <div>
          <label htmlFor="email">이메일</label>
          <input
            id="email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />   
        </div>
        <div>
          <label htmlFor="password">비밀번호</label>
          <input
            id="password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />   
        </div>

        {errorMsg && <p className="error-msg">{errorMsg}</p>}

        <button type="submit" disabled={isLoading}>
          {isLoading ? '로그인 중...' : '로그인'}
        </button>    
      </form>  
    </div>
  );
}

export default Login;
