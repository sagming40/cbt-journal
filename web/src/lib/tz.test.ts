import { describe, expect, it } from 'vitest';
import {
	formatAgo,
	formatDateTime,
	formatLocalInput,
	formatTime,
	fromLocalInput,
	toLocalInput
} from './tz.js';

describe('formatTime · toLocalInput', () => {
	it('설정한 시간대 기준으로 보인다 (기기 시간대와 상관없이)', () => {
		// 15:30Z = 서울 다음 날 00:30 · 뉴욕 10:30 (서버 테스트 TestCheckinDayFollowsTimezone과 같은 시각)
		// 뉴욕은 이날 06:00Z에 서머타임이 끝나 EST(-5)라 10:30이다(EDT였으면 11:30)
		expect(formatTime('2026-11-01T15:30:00Z', 'Asia/Seoul')).toBe('00:30');
		expect(formatTime('2026-11-01T15:30:00Z', 'America/New_York')).toBe('10:30');
		expect(toLocalInput('2026-11-01T15:30:00Z', 'Asia/Seoul')).toBe('2026-11-02T00:30');
	});

	it('자정은 24가 아니라 00', () => {
		expect(formatTime('2026-11-01T15:00:00Z', 'Asia/Seoul')).toBe('00:00');
	});
});

describe('fromLocalInput', () => {
	it('벽시계 → 실제 순간', () => {
		expect(fromLocalInput('2026-11-02T00:30', 'Asia/Seoul').toISOString()).toBe(
			'2026-11-01T15:30:00.000Z'
		);
	});

	it('toLocalInput과 왕복하면 그대로', () => {
		for (const iso of ['2026-01-15T03:07:00Z', '2026-07-01T23:59:00Z']) {
			for (const tz of ['Asia/Seoul', 'America/New_York', 'Europe/London']) {
				expect(fromLocalInput(toLocalInput(iso, tz), tz).toISOString()).toBe(
					new Date(iso).toISOString()
				);
			}
		}
	});

	it('서머타임이 끝나는 날(뉴욕 2026-11-01, 01:00~02:00이 두 번)도 맞는다', () => {
		// 00:30 EDT(-4) = 04:30Z, 23:30 EST(-5) = 2026-11-02 04:30Z
		expect(fromLocalInput('2026-11-01T00:30', 'America/New_York').toISOString()).toBe(
			'2026-11-01T04:30:00.000Z'
		);
		expect(fromLocalInput('2026-11-01T23:30', 'America/New_York').toISOString()).toBe(
			'2026-11-02T04:30:00.000Z'
		);
		// 두 번 있는 01:30은 앞의 것(EDT)
		expect(fromLocalInput('2026-11-01T01:30', 'America/New_York').toISOString()).toBe(
			'2026-11-01T05:30:00.000Z'
		);
	});

	it('서머타임이 시작하는 날(뉴욕 2026-03-08, 02:00~03:00이 없음)', () => {
		expect(fromLocalInput('2026-03-08T01:30', 'America/New_York').toISOString()).toBe(
			'2026-03-08T06:30:00.000Z'
		);
		expect(fromLocalInput('2026-03-08T03:30', 'America/New_York').toISOString()).toBe(
			'2026-03-08T07:30:00.000Z'
		);
	});

	it('형식이 틀리면 오류', () => {
		expect(() => fromLocalInput('2026-11-01 00:30', 'Asia/Seoul')).toThrow();
	});
});

describe('formatLocalInput', () => {
	it('벽시계 값을 그대로 말로 (기기 시간대와 상관없이)', () => {
		expect(formatLocalInput('2026-11-02T00:30')).toBe('2026년 11월 2일 (월) 00:30');
		expect(formatLocalInput('2026-03-01T23:05')).toBe('2026년 3월 1일 (일) 23:05');
	});
});

describe('formatDateTime · formatAgo', () => {
	it('설정한 시간대의 날짜 · 요일로 보인다', () => {
		// 15:30Z = 서울 11월 2일(월) 00:30 · 뉴욕 11월 1일(일) 10:30
		expect(formatDateTime('2026-11-01T15:30:00Z', 'Asia/Seoul')).toBe('11월 2일 (월) 00:30');
		expect(formatDateTime('2026-11-01T15:30:00Z', 'America/New_York')).toBe('11월 1일 (일) 10:30');
	});

	it('지난 시간을 말로', () => {
		const now = new Date('2026-11-02T12:00:00Z');
		expect(formatAgo('2026-11-02T11:59:30Z', now)).toBe('방금');
		expect(formatAgo('2026-11-02T11:57:00Z', now)).toBe('3분 전');
		expect(formatAgo('2026-11-02T10:00:00Z', now)).toBe('2시간 전');
		expect(formatAgo('2026-10-31T12:00:00Z', now)).toBe('2일 전');
	});
});
