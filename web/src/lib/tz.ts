// 사용자 시간대(설정의 users.timezone) 기준으로 날짜 · 시각을 다룬다.
// 브라우저(기기)의 시간대가 아니라 설정한 시간대가 기준이다 — 서버가 "그날"을 계산하는 기준과 같게 (DATA_MODEL 3.4절).
// 비유: 해외여행 중인 휴대폰 시계가 아니라, 집에 걸어 둔 시계를 본다.

/** 시간대 tz에서 본 date의 벽시계 값 (연 · 월 · 일 · 시 · 분 · 초) */
function wallClock(date: Date, tz: string) {
	const parts = new Intl.DateTimeFormat('en-US', {
		timeZone: tz,
		hourCycle: 'h23',
		year: 'numeric',
		month: '2-digit',
		day: '2-digit',
		hour: '2-digit',
		minute: '2-digit',
		second: '2-digit'
	}).formatToParts(date);
	const get = (type: string) => Number(parts.find((p) => p.type === type)?.value);
	return {
		year: get('year'),
		month: get('month'),
		day: get('day'),
		hour: get('hour'),
		minute: get('minute'),
		second: get('second')
	};
}

const pad = (n: number) => String(n).padStart(2, '0');

/** 시각을 tz 기준 "HH:mm"으로 (목록 왼쪽 여백의 시각) */
export function formatTime(iso: string | Date, tz: string): string {
	const w = wallClock(new Date(iso), tz);
	return `${pad(w.hour)}:${pad(w.minute)}`;
}

/** 시각을 <input type="datetime-local">에 넣을 tz 기준 "YYYY-MM-DDTHH:mm"으로 */
export function toLocalInput(iso: string | Date, tz: string): string {
	const w = wallClock(new Date(iso), tz);
	return `${w.year}-${pad(w.month)}-${pad(w.day)}T${pad(w.hour)}:${pad(w.minute)}`;
}

/**
 * tz 기준 벽시계 "YYYY-MM-DDTHH:mm"을 실제 순간(Date)으로 바꾼다.
 * 그 시간대가 UTC와 몇 분 차이인지(오프셋)를 그 순간 기준으로 구해 뺀다. 서머타임 경계에서는 오프셋이 바뀌므로
 * 한 번 더 구해 맞춘다 — 없는 시각(봄에 건너뛰는 1시간)은 앞으로 밀리고, 두 번 있는 시각(가을)은 앞의 것을 고른다.
 */
export function fromLocalInput(value: string, tz: string): Date {
	const m = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})$/.exec(value);
	if (!m) throw new Error(`시각 형식이 아니에요: ${value}`);
	const [, y, mo, d, h, mi] = m.map(Number);
	const asUtc = Date.UTC(y, mo - 1, d, h, mi);
	const offset = (at: number) => {
		const w = wallClock(new Date(at), tz);
		return Date.UTC(w.year, w.month - 1, w.day, w.hour, w.minute, w.second) - at;
	};
	let guess = asUtc - offset(asUtc);
	const corrected = asUtc - offset(guess);
	if (corrected !== guess) guess = corrected;
	return new Date(guess);
}
