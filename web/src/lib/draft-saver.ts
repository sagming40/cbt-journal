// 초안 자동 저장 (DESIGN 4.3절 "자동 저장" · "저장 순서", DATA_MODEL 4.3절). 화면과 떼어 놓아 테스트한다.
// 저장 요청은 한 번에 하나만 보낸다 — 겹치면 앞 요청이 올린 rev를 뒤 요청이 몰라서 한 기기가 자기 자신과 409를 낸다.
// 그동안 쓴 내용은 마지막 것만 모아 다음에 보낸다.
// 비유: 우체부가 한 명뿐이다. 다녀오는 동안 쓴 편지는 가장 새것만 다음 배달에 들려 보낸다.

/** 저장 한 번의 결과. ok가 아니면 내용은 화면에 그대로 남고 pending으로 돌아간다 */
export type SaveResult =
	| { kind: 'ok'; id: string; rev: number }
	| { kind: 'conflict' } // 다른 기기가 먼저 고침 (409 DRAFT_CONFLICT) → D-02
	| { kind: 'missing' } // 초안이 없어짐 (404)
	| { kind: 'limit'; message: string } // 초안 10개 (409 DRAFT_LIMIT)
	| { kind: 'failed'; message: string }; // 연결 끊김 · 503 · 그 밖의 오류

export type SaveState =
	| 'idle' // 저장할 것이 아직 없음
	| 'saving' // 입력을 기다리거나 보내는 중
	| 'saved'
	| 'failed'
	| 'limit'
	| 'conflict'
	| 'missing';

type Options<P> = {
	/** 첫 저장 */
	create: (payload: P) => Promise<SaveResult>;
	/** 그 뒤 저장. rev는 마지막으로 받은 값 */
	update: (id: string, rev: number, payload: P) => Promise<SaveResult>;
	/** 입력이 멈춘 뒤 기다리는 시간 */
	delayMs?: number;
	onState?: (state: SaveState, info: { message: string; savedAt: Date | null }) => void;
	/** 첫 저장으로 초안 id가 생겼을 때 (주소에 draft를 넣는다) */
	onCreated?: (id: string) => void;
	now?: () => Date;
};

export class DraftSaver<P> {
	id: string | null = null;
	rev = 0;
	state: SaveState = 'idle';
	message = '';
	savedAt: Date | null = null;

	private pending: P | null = null;
	private timer: ReturnType<typeof setTimeout> | null = null;
	private running: Promise<void> | null = null;
	private readonly delayMs: number;

	constructor(private readonly opts: Options<P>) {
		this.delayMs = opts.delayMs ?? 1500;
	}

	/** 이미 있는 초안을 이어 쓴다 */
	resume(id: string, rev: number, savedAt: Date) {
		this.id = id;
		this.rev = rev;
		this.savedAt = savedAt;
		this.setState('saved');
	}

	/** 입력이 바뀜. 멈추면 저장한다 */
	change(payload: P) {
		this.pending = payload;
		this.clearTimer();
		// 사용자의 답이 필요한 상태에서는 저절로 보내지 않는다 — 답한 뒤 마지막 내용이 나간다
		if (this.blocked) return;
		this.setState('saving');
		this.timer = setTimeout(() => {
			this.timer = null;
			void this.run();
		}, this.delayMs);
	}

	/** 기다리지 않고 지금 보내고, 남은 것이 없을 때까지 기다린다 (단계 이동 · 나가기 · 완료 전) */
	async flush(): Promise<SaveState> {
		this.clearTimer();
		await this.run();
		return this.state;
	}

	/** [다시 시도] */
	retry() {
		return this.flush();
	}

	/** D-02 "지금 내용으로 덮어쓰기": 최신 rev를 받아 그 번호로 저장한다 (강제 저장 플래그가 없다, DATA_MODEL 4.3절) */
	overwrite(latestRev: number, payload: P) {
		this.rev = latestRev;
		this.pending = payload;
		this.setState('saving');
		return this.flush();
	}

	/** D-02 "그쪽 내용 불러오기": 서버 내용을 받아들인다. 쌓인 입력은 버린다 */
	adopt(rev: number, savedAt: Date) {
		this.clearTimer();
		this.pending = null;
		this.rev = rev;
		this.savedAt = savedAt;
		this.message = '';
		this.setState('saved');
	}

	/** "초안이 없어짐" → [새 초안으로 저장]: 지금 내용으로 초안을 새로 만든다 */
	startOver(payload: P) {
		this.id = null;
		this.rev = 0;
		this.pending = payload;
		this.setState('saving');
		return this.flush();
	}

	/** 충돌을 풀려고 읽어 보니 초안이 없어졌을 때 */
	markMissing() {
		this.setState('missing');
	}

	/** 아직 서버에 안 간 내용이 있는가 */
	get hasUnsaved() {
		return this.pending !== null || this.running !== null || this.timer !== null;
	}

	/** 사용자의 답이 있어야 다음으로 가는 상태 */
	get blocked() {
		return this.state === 'conflict' || this.state === 'missing';
	}

	private clearTimer() {
		if (this.timer !== null) {
			clearTimeout(this.timer);
			this.timer = null;
		}
	}

	private setState(state: SaveState) {
		this.state = state;
		this.opts.onState?.(state, { message: this.message, savedAt: this.savedAt });
	}

	private run(): Promise<void> {
		// 이미 도는 중이면 같이 기다린다. 그 안의 반복이 pending을 마저 보낸다.
		this.running ??= this.loop().finally(() => {
			this.running = null;
		});
		return this.running;
	}

	private async loop() {
		while (this.pending !== null) {
			const payload = this.pending;
			this.pending = null;
			this.setState('saving');
			const wasNew = this.id === null;
			let result: SaveResult;
			try {
				result =
					this.id === null
						? await this.opts.create(payload)
						: await this.opts.update(this.id, this.rev, payload);
			} catch {
				result = { kind: 'failed', message: '연결이 안 돼요' };
			}
			if (result.kind === 'ok') {
				this.id = result.id;
				this.rev = result.rev;
				this.savedAt = (this.opts.now ?? (() => new Date()))();
				this.message = '';
				if (wasNew) this.opts.onCreated?.(result.id);
				continue;
			}
			// 실패: 그사이 더 새 입력이 없으면 이 내용을 다시 들고 멈춘다. 내용을 버리지 않는다.
			this.pending ??= payload;
			this.message = 'message' in result ? result.message : '';
			this.setState(result.kind);
			return;
		}
		this.setState(this.savedAt ? 'saved' : 'idle');
	}
}
