import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { DraftSaver, type SaveResult } from './draft-saver.js';

type Call = { kind: 'create' | 'update'; rev?: number; payload: string };

// 서버 흉내: 응답을 테스트가 손으로 놓아 주어 "요청이 도는 중"을 만들 수 있다.
function setup(delayMs = 1500) {
	const calls: Call[] = [];
	const replies: ((r: SaveResult) => void)[] = [];
	const wait = () => new Promise<SaveResult>((resolve) => replies.push(resolve));
	const created: string[] = [];
	const saver = new DraftSaver<string>({
		delayMs,
		create: (payload) => {
			calls.push({ kind: 'create', payload });
			return wait();
		},
		update: (_id, rev, payload) => {
			calls.push({ kind: 'update', rev, payload });
			return wait();
		},
		onCreated: (id) => created.push(id)
	});
	return { saver, calls, replies, created };
}

const tick = () => vi.advanceTimersByTimeAsync(0);

describe('DraftSaver', () => {
	beforeEach(() => vi.useFakeTimers());
	afterEach(() => vi.useRealTimers());

	it('입력이 멈춘 뒤에 마지막 내용만 저장한다', async () => {
		const { saver, calls } = setup();
		saver.change('가');
		await vi.advanceTimersByTimeAsync(1000);
		saver.change('가나');
		await vi.advanceTimersByTimeAsync(1499);
		expect(calls).toEqual([]);
		expect(saver.state).toBe('saving');
		await vi.advanceTimersByTimeAsync(1);
		expect(calls).toEqual([{ kind: 'create', payload: '가나' }]);
	});

	it('첫 저장은 한 번만 만들고 그 뒤에는 받은 rev로 저장한다', async () => {
		const { saver, calls, replies, created } = setup();
		saver.change('가');
		await vi.advanceTimersByTimeAsync(1500);
		// 첫 요청이 도는 동안 더 쓴다 — 초안이 둘 생기면 안 된다
		saver.change('가나');
		await vi.advanceTimersByTimeAsync(1500);
		saver.change('가나다');
		await vi.advanceTimersByTimeAsync(1500);
		expect(calls).toHaveLength(1);

		replies[0]({ kind: 'ok', id: 'd1', rev: 1 });
		await tick();
		expect(created).toEqual(['d1']);
		// 그동안 쓴 것은 마지막 것만, rev 1로
		expect(calls[1]).toEqual({ kind: 'update', rev: 1, payload: '가나다' });

		replies[1]({ kind: 'ok', id: 'd1', rev: 2 });
		await tick();
		expect(saver.state).toBe('saved');
		expect(saver.rev).toBe(2);
		expect(saver.hasUnsaved).toBe(false);
		expect(calls).toHaveLength(2);
	});

	it('저장 요청은 겹치지 않는다 (자기 자신과 409를 내지 않게)', async () => {
		const { saver, calls, replies } = setup();
		saver.resume('d1', 5, new Date());
		saver.change('가');
		const flushed = saver.flush();
		saver.change('가나');
		void saver.flush();
		await tick();
		expect(calls).toEqual([{ kind: 'update', rev: 5, payload: '가' }]);
		replies[0]({ kind: 'ok', id: 'd1', rev: 6 });
		await tick();
		expect(calls[1]).toEqual({ kind: 'update', rev: 6, payload: '가나' });
		replies[1]({ kind: 'ok', id: 'd1', rev: 7 });
		expect(await flushed).toBe('saved');
	});

	it('flush는 기다리지 않고 보내고 끝날 때까지 기다린다', async () => {
		const { saver, calls, replies } = setup();
		saver.resume('d1', 1, new Date());
		saver.change('다음 단계');
		const flushed = saver.flush();
		await tick();
		expect(calls).toHaveLength(1);
		replies[0]({ kind: 'ok', id: 'd1', rev: 2 });
		expect(await flushed).toBe('saved');
	});

	it('실패하면 내용을 들고 멈추고, 다시 시도하면 그 내용을 보낸다', async () => {
		const { saver, calls, replies } = setup();
		saver.resume('d1', 1, new Date());
		saver.change('글');
		await vi.advanceTimersByTimeAsync(1500);
		replies[0]({ kind: 'failed', message: '지금은 초안을 저장할 수 없어요' });
		await tick();
		expect(saver.state).toBe('failed');
		expect(saver.message).toBe('지금은 초안을 저장할 수 없어요');
		expect(saver.hasUnsaved).toBe(true);

		void saver.retry();
		await tick();
		expect(calls[1]).toEqual({ kind: 'update', rev: 1, payload: '글' });
		replies[1]({ kind: 'ok', id: 'd1', rev: 2 });
		await tick();
		expect(saver.state).toBe('saved');
		expect(saver.message).toBe('');
	});

	it('네트워크 오류(예외)도 실패로 친다', async () => {
		const saver = new DraftSaver<string>({
			create: () => Promise.reject(new Error('끊김')),
			update: () => Promise.reject(new Error('끊김'))
		});
		saver.change('글');
		expect(await saver.flush()).toBe('failed');
		expect(saver.hasUnsaved).toBe(true);
	});

	it('실패 뒤 다시 입력하면 저절로 다시 저장한다', async () => {
		const { saver, calls, replies } = setup();
		saver.resume('d1', 1, new Date());
		saver.change('글');
		await vi.advanceTimersByTimeAsync(1500);
		replies[0]({ kind: 'failed', message: '연결이 안 돼요' });
		await tick();
		saver.change('글을 더');
		await vi.advanceTimersByTimeAsync(1500);
		expect(calls[1].payload).toBe('글을 더');
	});

	it('충돌하면 사용자가 답할 때까지 저절로 보내지 않는다', async () => {
		const { saver, calls, replies } = setup();
		saver.resume('d1', 1, new Date());
		saver.change('내 글');
		await vi.advanceTimersByTimeAsync(1500);
		replies[0]({ kind: 'conflict' });
		await tick();
		expect(saver.state).toBe('conflict');
		saver.change('내 글 더');
		await vi.advanceTimersByTimeAsync(5000);
		expect(calls).toHaveLength(1);
		expect(saver.state).toBe('conflict');

		// 덮어쓰기: 최신 rev로
		const done = saver.overwrite(7, '내 글 더');
		await tick();
		expect(calls[1]).toEqual({ kind: 'update', rev: 7, payload: '내 글 더' });
		replies[1]({ kind: 'ok', id: 'd1', rev: 8 });
		expect(await done).toBe('saved');
	});

	it('그쪽 내용 불러오기는 쌓인 입력을 버린다', async () => {
		const { saver, calls, replies } = setup();
		saver.resume('d1', 1, new Date());
		saver.change('내 글');
		await vi.advanceTimersByTimeAsync(1500);
		replies[0]({ kind: 'conflict' });
		await tick();
		saver.adopt(9, new Date());
		expect(saver.state).toBe('saved');
		expect(saver.hasUnsaved).toBe(false);
		expect(saver.rev).toBe(9);
		await vi.advanceTimersByTimeAsync(5000);
		expect(calls).toHaveLength(1);
	});

	it('초안이 없어지면 새 초안으로 저장할 수 있다', async () => {
		const { saver, calls, replies, created } = setup();
		saver.resume('d1', 3, new Date());
		saver.change('글');
		await vi.advanceTimersByTimeAsync(1500);
		replies[0]({ kind: 'missing' });
		await tick();
		expect(saver.state).toBe('missing');

		const done = saver.startOver('글');
		await tick();
		expect(calls[1]).toEqual({ kind: 'create', payload: '글' });
		replies[1]({ kind: 'ok', id: 'd2', rev: 1 });
		expect(await done).toBe('saved');
		expect(created).toEqual(['d2']);
		expect(saver.id).toBe('d2');
	});

	it('10개 제한에 걸려도 내용을 들고 있고, 자리가 나면 다시 만든다', async () => {
		const { saver, calls, replies } = setup();
		saver.change('글');
		await vi.advanceTimersByTimeAsync(1500);
		replies[0]({ kind: 'limit', message: '초안은 10개까지' });
		await tick();
		expect(saver.state).toBe('limit');
		expect(saver.id).toBeNull();
		expect(saver.hasUnsaved).toBe(true);
		void saver.retry();
		await tick();
		expect(calls[1]).toEqual({ kind: 'create', payload: '글' });
	});
});
