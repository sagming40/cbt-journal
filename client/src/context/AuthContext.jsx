// client/src/context/AuthContext.jsx

import { createContext, useState, useContext, children } from "react";

// 1. 사물함 자체를 만든다
const AuthContext = createContext(null);

// 2. 사물함을 관리하는 관리인 컴포넌트
export const AuthProvider = ({ children }) => {
  // 새로고침해도 로그인이 풀리지 않도록, localStorage에 저장된 값을 초기값으로 사용
  const [token, setToken] = useState(localStorage.getItem('token'));
  const [user, setUser] = useState(() => {
    const savedUser = localStorage.getItem('user');
    return savedUser ? JSON.parse(savedUser) : null;
  });
  
  // 로그인 성공 시 호출할 함수
  const login = (newToken, newUser) => {
    localStorage.setItem('token', newToken);
    localStorage.setItem('user', JSON.stringify(newUser));
    setToken(newToken);
    setUser(newUser);
  };

  // 로그아웃 시 호출할 함수
  const logout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    setToken(null);
    setUser(null);
  };

  return (
    <AuthContext.Provider value={{ token, user, login, logout }}>
        {children}
    </AuthContext.Provider>
  );
};

// 3. 다른 컴포넌트에서 사물함 내용물을 쉽게 꺼내 쓰기 위한 도구
export const useAuth = () => useContext(AuthContext);
