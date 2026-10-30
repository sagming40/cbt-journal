// 로그인 뒤 돌아갈 주소(?next=)를 앱 안 경로로만 받는다 — 바깥 주소로 튕기는 오픈 리다이렉트 방지 (DESIGN 3.1절, ARCHITECTURE 4절).
// 비유: "돌아갈 곳" 쪽지에 우리 건물 안 방 번호만 적게 한다. 다른 건물 주소가 적혀 있으면 로비(/)로 보낸다.

// 주소를 해석할 가짜 출처. 해석한 결과의 출처가 이것과 다르면 바깥 주소였다는 뜻이다.
const BASE = 'http://app.invalid';

export function safeNext(raw: string | null | undefined): string {
	if (!raw) return '/';
	// /로 시작해야 하고, //evil.com(프로토콜 생략 주소)이나 /\evil.com(브라우저가 //처럼 읽음)은 안 된다.
	if (!raw.startsWith('/') || raw.startsWith('//') || raw.startsWith('/\\')) return '/';
	// 브라우저는 주소 안의 탭 · 줄바꿈을 지우고 해석한다: "/\t/evil.com"이 "//evil.com"이 된다. 제어 문자가 있으면 거절.
	// eslint-disable-next-line no-control-regex -- 제어 문자를 찾는 게 목적이다
	if (/[\u0000-\u001f\u007f]/.test(raw)) return '/';
	// 마지막으로 실제 해석 결과의 출처가 같은지 본다 (위 규칙이 놓친 꼴까지).
	let url: URL;
	try {
		url = new URL(raw, BASE);
	} catch {
		return '/';
	}
	if (url.origin !== BASE) return '/';
	return url.pathname + url.search + url.hash;
}
