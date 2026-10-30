// 로그인 상태. 세션 토큰은 HttpOnly 쿠키에만 있고 자바스크립트는 못 읽는다 (CLAUDE.md "하지 말 것").
// 여기에는 화면에 보여 줄 "내 정보"만 둔다 — 서버에 GET /api/me로 물어서 채운다.
import type { components } from '#lib/api/schema.js';

export type Me = components['schemas']['Me'];

class AuthState {
	// $state: 값이 바뀌면 이 값을 쓰는 화면이 다시 그려진다.
	user = $state<Me | null>(null);

	set(user: Me) {
		this.user = user;
	}

	clear() {
		this.user = null;
	}
}

export const auth = new AuthState();
