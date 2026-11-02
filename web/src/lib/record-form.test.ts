import { describe, expect, it } from 'vitest';
import {
	canFinish,
	clampStep,
	emptyForm,
	fillAfter,
	firstErrorStep,
	furthestStep,
	isBlank,
	stepComplete,
	type RecordForm
} from './record-form.js';

const now = new Date('2026-11-02T03:00:00Z');
const em = (id: number, before: number | null, after: number | null) => ({
	emotion_id: id,
	intensity_before: before,
	intensity_after: after
});

function full(): RecordForm {
	return {
		...emptyForm(now),
		situation: '회의',
		emotions: [em(1, 80, 35)],
		automatic_thought: '또 망쳤다',
		alternative_thought: '끝까지 해냈다'
	};
}

describe('필수 칸', () => {
	it('공백만 있으면 안 쓴 것', () => {
		expect(stepComplete(1, { ...emptyForm(now), situation: '   ' })).toBe(false);
		expect(stepComplete(1, { ...emptyForm(now), situation: '회의' })).toBe(true);
	});

	it('4 · 5 · 6단계는 비워도 된다', () => {
		const f = emptyForm(now);
		expect([4, 5, 6].map((s) => stepComplete(s, f))).toEqual([true, true, true]);
	});

	it('8단계는 감정마다 전 · 후가 다 있어야 한다', () => {
		expect(stepComplete(8, { ...full(), emotions: [em(1, 80, null)] })).toBe(false);
		expect(stepComplete(8, { ...full(), emotions: [] })).toBe(false);
		expect(stepComplete(8, full())).toBe(true);
	});

	it('[완료]는 전부 찼을 때만', () => {
		expect(canFinish(full())).toBe(true);
		expect(canFinish({ ...full(), alternative_thought: '' })).toBe(false);
		expect(canFinish({ ...full(), evidence_for: '' })).toBe(true);
	});
});

describe('주소의 step 보정', () => {
	it('빈 초안에 step=5면 1단계로', () => {
		expect(clampStep(5, emptyForm(now))).toBe(1);
	});

	it('범위 밖 · 숫자 아님', () => {
		expect(clampStep(0, full())).toBe(1);
		expect(clampStep(-3, full())).toBe(1);
		expect(clampStep(NaN, full())).toBe(1);
		expect(clampStep(2.5, full())).toBe(1);
		expect(clampStep(99, full())).toBe(8);
	});

	it('휴대폰에서 3단계까지 쓴 초안은 PC에서 4 ~ 7단계까지 열린다', () => {
		const f = { ...full(), alternative_thought: '' };
		expect(furthestStep(f)).toBe(7);
		expect(clampStep(4, f)).toBe(4);
		expect(clampStep(8, f)).toBe(7);
	});

	it('앞 단계를 비우면 그 뒤로는 못 간다', () => {
		expect(clampStep(6, { ...full(), automatic_thought: '' })).toBe(3);
	});
});

describe('후 강도 처음 값', () => {
	it('비어 있으면 전 값에서 시작하고, 매긴 값은 그대로', () => {
		expect(fillAfter([em(1, 80, null), em(2, 60, 20)])).toEqual([em(1, 80, 80), em(2, 60, 20)]);
	});
});

describe('빈 초안', () => {
	it('시각만 있으면 빈 것 — 빈 초안을 만들지 않는다', () => {
		expect(isBlank(emptyForm(now))).toBe(true);
		expect(isBlank({ ...emptyForm(now), situation: '가' })).toBe(false);
		expect(isBlank({ ...emptyForm(now), emotions: [em(1, 50, null)] })).toBe(false);
		expect(isBlank({ ...emptyForm(now), distortion_ids: [3] })).toBe(false);
	});
});

describe('완료 오류가 난 칸으로 보내기', () => {
	it('가장 앞 단계', () => {
		expect(firstErrorStep({ alternative_thought: 'x', automatic_thought: 'x' })).toBe(3);
		expect(firstErrorStep({ occurred_at: 'x' })).toBe(1);
		expect(firstErrorStep({ draft_id: 'x' })).toBeNull();
		expect(firstErrorStep({})).toBeNull();
	});
});
