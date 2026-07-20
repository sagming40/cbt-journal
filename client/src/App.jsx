// client/src/App.jsx
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import DiaryListPage from './pages/DiaryListPage';
import DiaryDetailPage from './pages/DiaryDetailPage';
import Login from './pages/Login';

// 로그인 안 한 사람이 목록/상세 페이지에 못 들어오게 막는 경비원 역할
const PrivateRoute = ({ children }) => {
  const { token } = useAuth();
  
  // 토큰이 없으면 로그인 페이지로 돌아감 (로그인 페이지는 다음 단계 구현 예정)
  if (!token) {
    return <Navigate to="/login" replace />;
  }

  return children;
};

function App() {
  return (
    <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route
              path="/diaries"
              element={
                <PrivateRoute>
                  <DiaryListPage />  
                </PrivateRoute>
               }
             />
             <Route
               path="/diaries/:id"
               element={
                 <PrivateRoute>
                   <DiaryDetailPage /> 
                 </PrivateRoute>
                }
              />
              {/* 아직 만들지 않은 페이지들 ㅡ 나중에 채울 자리 */}
              <Route path="/login" element={<Login />} />
              <Route path="*" element={<Navigate to="/diaries" replace />} />   
            </Routes>  
        </BrowserRouter>
    </AuthProvider>
  );  
}

export default App;
