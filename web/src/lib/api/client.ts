// API 호출은 이 클라이언트로만 한다 (decisions 0003).
// 경로와 응답 타입은 api/openapi.yaml에서 생성한 schema.d.ts에서 온다 — 계약서에 없는 경로는 타입 오류가 난다.
import createClient, { type Middleware } from 'openapi-fetch';
import { goto } from '$app/navigation';
import { resolve } from '$app/paths';
import { page } from '$app/state';
import { auth } from '#lib/auth.svelte.js';
import type { paths } from './schema.js';

// 로그인 · 가입 요청의 401은 "비밀번호가 틀림"이다 — 그 화면에서 문구로 보여 준다.
// 이걸 빼지 않으면 로그인 화면에서 로그인 화면으로 튕긴다 (ARCHITECTURE 4절).
const CREDENTIAL_PATHS = new Set(['/api/auth/login', '/api/auth/signup']);

// 401을 한 곳에서 처리한다: 로그인 상태를 지우고 지금 주소를 기억한 채 로그인 화면으로 (DESIGN 7장).
// 화면마다 따로 처리하지 않는다.
const onUnauthorized: Middleware = {
	async onResponse({ request, response }) {
		if (response.status !== 401) return;
		if (CREDENTIAL_PATHS.has(new URL(request.url).pathname)) return;
		auth.clear();
		if (page.url.pathname === resolve('/login')) return;
		const here = page.url.pathname + page.url.search;
		await goto(resolve(`/login?next=${encodeURIComponent(here)}`), { replace: true });
	}
};

// baseUrl을 비우면 지금 페이지와 같은 출처로 요청한다: 개발은 Vite 프록시, 운영은 같은 Go 서버.
export const api = createClient<paths>({ baseUrl: '' });
api.use(onUnauthorized);
