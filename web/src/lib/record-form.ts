// S-04 사고 기록 작성의 단계 규칙 (DESIGN 4.3절). 화면과 떼어 놓아 테스트한다.
// 규칙의 원본은 서버다(decisions 0008). 여기서는 "버튼을 켤지 · 어느 단계로 보낼지"만 정한다.
import type { components } from '#lib/api/schema.js';

type RecordEmotionInput = components['schemas']['RecordEmotionInput'];
type DraftInput = components['schemas']['DraftInput'];
type RecordInput = components['schemas']['RecordInput'];

export const LAST_STEP = 8;

/** 작성 중인 기록 전체. 초안 · 완료 · 수정이 같은 모양을 쓴다 */
export type RecordForm = Omit<DraftInput, 'step'>;

/**
 * 단계 제목 · 도움말. **문구는 초안이다** — 사용자가 다듬는다 (DESIGN 4.3절, PLAN M5).
 * 칸 이름은 REQUIREMENTS 용어를 쓴다.
 */
export const STEPS: { title: string; help: string }[] = [
	{ title: '어떤 일이 있었나요?', help: '언제, 어디서, 무슨 일이었는지 있었던 그대로 적어 봐요.' },
	{ title: '그때 어떤 감정이 들었나요?', help: '여러 개 골라도 돼요. 그때 느낀 세기로 매겨 봐요.' },
	{ title: '그때 머릿속에 스친 생각은?', help: '떠오른 그대로 써도 돼요.' },
	{
		title: '이 생각에 어떤 왜곡이 섞여 있을까요?',
		help: '찾지 못해도 괜찮아요. 비워 두고 넘어가도 돼요.'
	},
	{
		title: '이 생각이 맞다고 볼 수 있는 사실은?',
		help: '느낌 말고 사실로 적어 봐요. 없으면 비워 둬도 돼요.'
	},
	{ title: '이 생각과 맞지 않는 사실은?', help: '작은 것이어도 괜찮아요. 없으면 비워 둬도 돼요.' },
	{ title: '조금 더 균형 잡힌 생각은?', help: '앞에서 쓴 사실들을 보고 다시 적어 봐요.' },
	{ title: '지금은 감정이 어떤가요?', help: '안 줄었어도 괜찮아요. 지금 느끼는 세기로 매겨 봐요.' }
];

export function emptyForm(now: Date): RecordForm {
	return {
		// 일어난 시각은 화면을 연 때로 정해 둔다 — 이틀 뒤 이어 써도 밀리지 않게 (DATA_MODEL 4.3절, T6)
		occurred_at: now.toISOString(),
		situation: '',
		automatic_thought: '',
		evidence_for: '',
		evidence_against: '',
		alternative_thought: '',
		emotions: [],
		distortion_ids: []
	};
}

/** 아직 아무것도 안 썼는가 — 빈 초안을 만들지 않는다 (DESIGN 4.3절 "첫 진입") */
export function isBlank(f: RecordForm): boolean {
	return (
		!f.situation &&
		!f.automatic_thought &&
		!f.evidence_for &&
		!f.evidence_against &&
		!f.alternative_thought &&
		f.emotions.length === 0 &&
		f.distortion_ids.length === 0
	);
}

const filled = (s: string) => s.trim().length > 0;

/** 그 단계의 필수 칸이 찼는가. 차지 않으면 [다음] 비활성 */
export function stepComplete(step: number, f: RecordForm): boolean {
	switch (step) {
		case 1:
			return filled(f.situation) && Boolean(f.occurred_at);
		case 2:
			return f.emotions.length > 0;
		case 3:
			return filled(f.automatic_thought);
		case 7:
			return filled(f.alternative_thought);
		case 8:
			return (
				f.emotions.length > 0 &&
				f.emotions.every((e) => e.intensity_before !== null && e.intensity_after !== null)
			);
		default:
			return true; // 4 · 5 · 6은 비워도 된다
	}
}

/** 들어갈 수 있는 가장 뒤 단계 = 앞 단계의 필수 칸이 처음으로 빈 곳 */
export function furthestStep(f: RecordForm): number {
	for (let step = 1; step < LAST_STEP; step++) {
		if (!stepComplete(step, f)) return step;
	}
	return LAST_STEP;
}

/** 주소의 step을 믿지 않는다: 범위 밖 · 숫자 아님 · 앞이 빈 단계는 들어갈 수 있는 곳으로 (DESIGN 4.3절 "이동") */
export function clampStep(requested: number, f: RecordForm): number {
	if (!Number.isInteger(requested) || requested < 1) return 1;
	return Math.min(requested, furthestStep(f));
}

/** 전부 찼는가 — 수정 화면의 [저장], 작성의 [완료] */
export function canFinish(f: RecordForm): boolean {
	return [1, 2, 3, 7, 8].every((step) => stepComplete(step, f));
}

/**
 * 8단계에 들어올 때 비어 있는 "후"를 "전" 값으로 채운다 (DESIGN 4.3절 "8단계 처음 값").
 * 50에서 시작하면 숫자를 한쪽으로 끈다. 이미 매긴 값은 그대로.
 */
export function fillAfter(emotions: RecordEmotionInput[]): RecordEmotionInput[] {
	return emotions.map((e) =>
		e.intensity_after === null ? { ...e, intensity_after: e.intensity_before } : e
	);
}

/** 서버가 준 칸 오류 → 그 칸이 있는 가장 앞 단계 (DESIGN 4.3절 "완료 오류") */
const FIELD_STEP: Record<string, number> = {
	situation: 1,
	occurred_at: 1,
	emotions: 2,
	automatic_thought: 3,
	distortion_ids: 4,
	evidence_for: 5,
	evidence_against: 6,
	alternative_thought: 7
};

export function firstErrorStep(fields: Record<string, string>): number | null {
	const steps = Object.keys(fields)
		.map((k) => FIELD_STEP[k])
		.filter((s) => s !== undefined);
	return steps.length > 0 ? Math.min(...steps) : null;
}

export function toRecordInput(f: RecordForm): RecordInput {
	return { ...f };
}

export function toDraftInput(f: RecordForm, step: number): DraftInput {
	return { ...f, step };
}

/** 서버 응답(초안 · 기록)에서 폼 칸만 뽑는다. 저장 여부를 비교할 때 칸 순서가 같아야 해서 여기서 맞춘다 */
export function pickForm(src: RecordForm): RecordForm {
	return {
		occurred_at: src.occurred_at,
		situation: src.situation,
		automatic_thought: src.automatic_thought,
		evidence_for: src.evidence_for,
		evidence_against: src.evidence_against,
		alternative_thought: src.alternative_thought,
		emotions: src.emotions.map((e) => ({
			emotion_id: e.emotion_id,
			intensity_before: e.intensity_before,
			intensity_after: e.intensity_after
		})),
		distortion_ids: [...src.distortion_ids]
	};
}
