// API 오류 봉투(ARCHITECTURE 3.2절)에서 화면에 보여 줄 것을 꺼낸다.
// 서버 message는 이미 화면용 해요체라(DESIGN 8장) 그대로 쓴다.
import type { components } from './schema.js';

type ApiError = components['schemas']['Error'];

export const NETWORK_ERROR = '연결이 안 돼요';
const FALLBACK = '잠시 문제가 생겼어요. 다시 시도해 주세요';

function isApiError(e: unknown): e is ApiError {
	return typeof e === 'object' && e !== null && 'error' in e;
}

export function errorMessage(e: unknown): string {
	return isApiError(e) ? e.error.message : FALLBACK;
}

export function fieldErrors(e: unknown): Record<string, string> {
	return isApiError(e) ? (e.error.fields ?? {}) : {};
}

// 429의 Retry-After(초)를 "3분 뒤" 같은 말로 (DESIGN 4.8절 "남은 시간").
export function retryAfterText(response: Response): string {
	const secs = Number(response.headers.get('Retry-After'));
	if (!Number.isFinite(secs) || secs <= 0) return '';
	if (secs < 60) return `${secs}초 뒤`;
	return `${Math.ceil(secs / 60)}분 뒤`;
}
