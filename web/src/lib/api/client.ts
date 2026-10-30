// API 호출은 이 클라이언트로만 한다 (decisions 0003).
// 경로와 응답 타입은 api/openapi.yaml에서 생성한 schema.d.ts에서 온다 — 계약서에 없는 경로는 타입 오류가 난다.
import createClient from 'openapi-fetch';
import type { paths } from './schema.js';

// baseUrl을 비우면 지금 페이지와 같은 출처로 요청한다: 개발은 Vite 프록시, 운영은 같은 Go 서버.
// 401을 한 곳에서 처리하는 미들웨어는 M2에서 여기에 붙인다 (ARCHITECTURE 4절).
export const api = createClient<paths>({ baseUrl: '' });
