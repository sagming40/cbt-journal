import { describe, expect, it } from 'vitest';
import { safeNext } from './safe-next.js';

describe('safeNext', () => {
	it('앱 안 경로는 그대로', () => {
		expect(safeNext('/settings')).toBe('/settings');
		expect(safeNext('/records?from=2026-10-01#top')).toBe('/records?from=2026-10-01#top');
	});

	it('비었으면 /', () => {
		expect(safeNext(null)).toBe('/');
		expect(safeNext(undefined)).toBe('/');
		expect(safeNext('')).toBe('/');
	});

	it.each([
		['https://evil.com', '절대 주소'],
		['//evil.com', '프로토콜 생략 주소'],
		['/\\evil.com', '역슬래시 (브라우저가 //처럼 읽음)'],
		['/\t/evil.com', '탭 (브라우저가 지우고 //로 읽음)'],
		['/\n/evil.com', '줄바꿈'],
		['javascript:alert(1)', '자바스크립트 주소'],
		['settings', '/로 시작하지 않음']
	])('바깥 주소 %s (%s)는 /', (raw) => {
		expect(safeNext(raw)).toBe('/');
	});
});
