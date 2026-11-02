<script lang="ts">
	// S-04 사고 기록 작성 (DESIGN 4.3절, REQUIREMENTS T1 ~ T3 · T6).
	// 입력이 멈추면 Valkey 초안에 저장한다(draft-saver.ts). 어느 기기에서든 ?draft=로 열면 멈춘 단계부터 이어 쓴다.
	import { onMount, untrack } from 'svelte';
	import { afterNavigate, beforeNavigate, goto } from '$app/navigation';
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { toast } from 'svelte-sonner';
	import { api } from '#lib/api/client.js';
	import { NETWORK_ERROR, errorMessage, fieldErrors } from '#lib/api/errors.js';
	import type { components } from '#lib/api/schema.js';
	import FormError from '#lib/components/FormError.svelte';
	import RecordStepFields from '#lib/components/RecordStepFields.svelte';
	import SaveStatus from '#lib/components/SaveStatus.svelte';
	import StepShell from '#lib/components/StepShell.svelte';
	import * as AlertDialog from '#lib/components/ui/alert-dialog/index.js';
	import { Button } from '#lib/components/ui/button/index.js';
	import { DraftSaver, type SaveResult, type SaveState } from '#lib/draft-saver.js';
	import {
		LAST_STEP,
		canFinish,
		clampStep,
		emptyForm,
		fillAfter,
		firstErrorStep,
		isBlank,
		pickForm,
		stepComplete,
		toDraftInput,
		toRecordInput,
		type RecordForm
	} from '#lib/record-form.js';

	type Emotion = components['schemas']['Emotion'];
	type Distortion = components['schemas']['Distortion'];
	type DraftInput = components['schemas']['DraftInput'];
	type Draft = components['schemas']['Draft'];

	let emotions = $state<Emotion[] | null>(null);
	let distortions = $state<Distortion[] | null>(null);
	let form = $state<RecordForm>(emptyForm(new Date()));
	let ready = $state(false);
	let loadError = $state('');
	let draftGone = $state(false);

	let saveState = $state<SaveState>('idle');
	let saveMessage = $state('');
	let savedAt = $state<Date | null>(null);
	let fields = $state<Record<string, string>>({});
	let message = $state('');
	let finishing = $state(false);
	let resolving = $state(false);
	// 나가려던 곳 — 저장 안 된 내용이 있을 때 확인을 받는 동안 들고 있는다
	let leaveTo = $state<URL | null>(null);
	let leaving = false;

	// 주소가 단계의 원본이다(뒤로 가기가 이전 단계로 가게). 다만 믿지 않고 들어갈 수 있는 곳으로 보정한다.
	// 이동할 때만 다시 읽는다 — 글을 쓰는 것만으로 단계가 바뀌면 안 된다(주소가 step=5인 빈 초안에 첫 글자를 치는 순간 등).
	let step = $state(1);
	afterNavigate(() => {
		if (ready) step = clampStep(Number(page.url.searchParams.get('step') ?? 1), form);
	});

	function urlFor(s: number, draftId: string | null) {
		const draft = draftId ? `draft=${draftId}&` : '';
		return `${resolve('/(app)/records/new')}?${draft}step=${s}`;
	}

	async function toResult(call: Promise<{ data?: Draft; error?: unknown; response: Response }>) {
		const { data, error, response } = await call;
		if (data) return { kind: 'ok', id: data.id, rev: data.rev } satisfies SaveResult;
		const code =
			typeof error === 'object' && error !== null && 'error' in error
				? (error as { error: { code: string } }).error.code
				: '';
		if (response.status === 404) return { kind: 'missing' } satisfies SaveResult;
		if (code === 'DRAFT_CONFLICT') return { kind: 'conflict' } satisfies SaveResult;
		const text = errorMessage(error);
		if (code === 'DRAFT_LIMIT') return { kind: 'limit', message: text } satisfies SaveResult;
		return { kind: 'failed', message: text } satisfies SaveResult;
	}

	const saver = new DraftSaver<DraftInput>({
		create: (body) => toResult(api.POST('/api/drafts', { body })),
		update: (id, rev, body) =>
			toResult(api.PUT('/api/drafts/{id}', { params: { path: { id } }, body: { ...body, rev } })),
		onState: (state, info) => {
			saveState = state;
			saveMessage = info.message;
			savedAt = info.savedAt;
		},
		// 주소에 초안 id를 넣는다. 쓰는 중이니 커서 · 스크롤을 건드리지 않는다.
		onCreated: (id) =>
			// eslint-disable-next-line svelte/no-navigation-without-resolve -- urlFor가 resolve로 만든 주소다
			goto(urlFor(step, id), { replace: true, reset: false })
	});

	// 내용(+ 멈춘 단계)이 바뀌면 저장을 건다. 비교해서 정말 바뀌었을 때만 — 불러온 직후에 저장이 나가지 않게.
	const payload = $derived(toDraftInput($state.snapshot(form) as RecordForm, step));
	let lastPayload = '';
	$effect(() => {
		const snap = JSON.stringify(payload);
		if (!ready || snap === lastPayload) return;
		lastPayload = snap;
		// 빈 초안을 만들지 않는다 (DESIGN 4.3절 "첫 진입")
		if (saver.id === null && isBlank(payload)) return;
		untrack(() => saver.change($state.snapshot(payload) as DraftInput));
	});

	function applyDraft(d: Draft) {
		form = pickForm(d);
		lastPayload = '';
	}

	async function load() {
		loadError = '';
		draftGone = false;
		const draftId = page.url.searchParams.get('draft');
		try {
			const [em, ds, dr] = await Promise.all([
				api.GET('/api/emotions'),
				api.GET('/api/distortions'),
				draftId ? api.GET('/api/drafts/{id}', { params: { path: { id: draftId } } }) : null
			]);
			if (dr?.response.status === 404) {
				draftGone = true;
				return;
			}
			if (!em.data || !ds.data || (dr && !dr.data)) {
				loadError = errorMessage(em.error ?? ds.error ?? dr?.error);
				return;
			}
			emotions = em.data.items;
			distortions = ds.data.items;
			let target = Number(page.url.searchParams.get('step') ?? 1);
			if (dr?.data) {
				applyDraft(dr.data);
				saver.resume(dr.data.id, dr.data.rev, new Date(dr.data.updated_at));
				// step을 주소에 안 줬으면 멈춘 단계부터 (DESIGN 4.3절 "나중에 이어 쓰기")
				if (!page.url.searchParams.has('step')) target = dr.data.step;
			}
			const start = clampStep(target, form);
			if (start === LAST_STEP) form.emotions = fillAfter(form.emotions);
			lastPayload = JSON.stringify(toDraftInput($state.snapshot(form) as RecordForm, start));
			step = start;
			ready = true;
			if (String(start) !== page.url.searchParams.get('step')) {
				// eslint-disable-next-line svelte/no-navigation-without-resolve -- urlFor가 resolve로 만든 주소다
				await goto(urlFor(start, draftId), { replace: true, reset: false });
			}
		} catch {
			loadError = NETWORK_ERROR;
		}
	}

	onMount(() => {
		load();
		// 휴대폰에서 다른 앱으로 넘어가면 바로 저장한다 — 그대로 앱이 닫혀도 남게
		const onHide = () => {
			if (document.visibilityState === 'hidden' && saver.hasUnsaved) void saver.flush();
		};
		document.addEventListener('visibilitychange', onHide);
		return () => document.removeEventListener('visibilitychange', onHide);
	});

	function go(s: number) {
		fields = {};
		message = '';
		if (s === LAST_STEP) form.emotions = fillAfter(form.emotions);
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- urlFor가 resolve로 만든 주소다
		goto(urlFor(s, saver.id)).then(() => {
			// 단계를 옮기면 기다리지 않고 저장한다 (DESIGN 4.3절 "저장 순서")
			if (saver.hasUnsaved) void saver.flush();
		});
	}

	// [나중에 이어 쓰기] (DESIGN 4.3절): 저장을 확인한 뒤 오늘로
	async function saveForLater() {
		const state = await saver.flush();
		if (state === 'saved') {
			leaving = true;
			toast.success('이어 쓰기에 남겨 뒀어요');
			await goto(resolve('/(app)'));
		}
	}

	// [완료]: 남은 저장을 먼저 마친다 — 다른 기기가 먼저 고쳤으면 여기서 D-02로 멈춰 그쪽 내용을 말없이 덮지 않는다
	// (DATA_MODEL 4.3절). 초안 저장소가 죽어 있어도(failed) 화면 내용으로 완료할 수 있다.
	async function finish() {
		if (finishing) return;
		finishing = true;
		fields = {};
		message = '';
		try {
			const state = await saver.flush();
			if (state === 'conflict' || state === 'missing') return;
			const { data, error } = await api.POST('/api/records', {
				body: {
					...toRecordInput($state.snapshot(form) as RecordForm),
					draft_id: saver.id ?? undefined
				}
			});
			if (data) {
				leaving = true;
				toast.success('기록했어요');
				await goto(resolve('/(app)/records/[id]', { id: data.id }), { replace: true });
				return;
			}
			fields = fieldErrors(error);
			const at = firstErrorStep(fields);
			if (at !== null) go(at);
			else message = errorMessage(error);
		} catch {
			message = NETWORK_ERROR;
		} finally {
			finishing = false;
		}
	}

	// ── D-02 충돌 · 초안 없어짐 ──
	async function loadTheirs() {
		resolving = true;
		try {
			const id = saver.id!;
			const { data, response } = await api.GET('/api/drafts/{id}', { params: { path: { id } } });
			if (data) {
				applyDraft(data);
				saver.adopt(data.rev, new Date(data.updated_at));
				lastPayload = JSON.stringify(
					toDraftInput($state.snapshot(form) as RecordForm, clampStep(data.step, form))
				);
				step = clampStep(data.step, form);
				// eslint-disable-next-line svelte/no-navigation-without-resolve -- urlFor가 resolve로 만든 주소다
				await goto(urlFor(step, id), { replace: true });
			} else if (response.status === 404) {
				saver.markMissing();
			}
		} finally {
			resolving = false;
		}
	}

	async function overwrite() {
		resolving = true;
		try {
			const id = saver.id!;
			const { data, response } = await api.GET('/api/drafts/{id}', { params: { path: { id } } });
			if (data) await saver.overwrite(data.rev, $state.snapshot(payload) as DraftInput);
			else if (response.status === 404) {
				saver.markMissing();
			}
		} finally {
			resolving = false;
		}
	}

	async function startOver() {
		resolving = true;
		try {
			await saver.startOver($state.snapshot(payload) as DraftInput);
		} finally {
			resolving = false;
		}
	}

	// ── 나가기: 저장 안 된 내용을 잃지 않게 (DESIGN 4.3절 "나가기") ──
	beforeNavigate((nav) => {
		if (leaving || !nav.to || nav.to.url.pathname === page.url.pathname) return;
		if (nav.type === 'leave') {
			// 탭 닫기 · 새로고침: 브라우저 기본 확인 창만 쓸 수 있다
			if (saver.hasUnsaved) nav.cancel();
			return;
		}
		if (!saver.hasUnsaved) return;
		nav.cancel();
		const to = nav.to.url;
		// 그냥 아직 안 보낸 것뿐이면 보내고 나간다. 실패할 때만 묻는다.
		saver.flush().then((state) => {
			if (state === 'saved') {
				leaving = true;
				// eslint-disable-next-line svelte/no-navigation-without-resolve -- 앱 안에서 누른 링크의 주소 그대로다
				goto(to);
			} else {
				leaveTo = to;
			}
		});
	});

	function leaveAnyway() {
		const to = leaveTo;
		leaveTo = null;
		leaving = true;
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- 앱 안에서 누른 링크의 주소 그대로다
		if (to) goto(to);
	}

	const nextDisabled = $derived(
		finishing || (step === LAST_STEP ? !canFinish(form) : !stepComplete(step, form))
	);
	const stopMessage = $derived(saveState === 'failed' || saveState === 'limit' ? saveMessage : '');
</script>

<svelte:head><title>사고 기록 · CBT 저널</title></svelte:head>

{#if draftGone}
	<div class="space-y-4 py-6">
		<FormError message="이 초안을 찾을 수 없어요. 다른 기기에서 완료했거나 지웠을 수 있어요." />
		<a href={resolve('/(app)')} class="text-accent-text underline-offset-4 hover:underline"
			>오늘로 돌아가기</a
		>
	</div>
{:else if ready && emotions && distortions}
	<StepShell
		heading="사고 기록"
		{step}
		{form}
		{emotions}
		{distortions}
		message={message || stopMessage}
		nextLabel={step === LAST_STEP ? (finishing ? '완료하는 중…' : '완료') : '다음'}
		{nextDisabled}
		onback={() => (step > 1 ? go(step - 1) : goto(resolve('/(app)')))}
		onprev={() => go(step - 1)}
		onnext={() => (step === LAST_STEP ? finish() : go(step + 1))}
		onjump={go}
	>
		{#snippet status()}
			<SaveStatus status={saveState} {savedAt} onretry={() => saver.retry()} />
		{/snippet}

		{#snippet later()}
			{#if step >= 3}
				<!-- 1 ~ 3단계가 붙잡기, 그 뒤는 나중에 어디서든 (DESIGN 1장 원칙 1 · 4.3절) -->
				<Button variant="ghost" class="min-h-12 px-3" onclick={saveForLater}
					>나중에 이어 쓰기</Button
				>
			{/if}
		{/snippet}
		<RecordStepFields bind:form {step} {emotions} {distortions} {fields} />
	</StepShell>
{:else if loadError}
	<div class="space-y-4">
		<FormError message={loadError} />
		<Button variant="outline" class="min-h-11 border-input" onclick={load}>다시 시도</Button>
	</div>
{:else}
	<p class="text-muted-foreground">불러오는 중…</p>
{/if}

<!-- D-02 충돌. Esc · 바깥 누르기로 닫지 않는다 — 어느 쪽을 남길지 답해야 저장이 이어진다 -->
<AlertDialog.Root open={saveState === 'conflict'}>
	<AlertDialog.Content
		class="rounded-md"
		escapeKeydownBehavior="ignore"
		interactOutsideBehavior="ignore"
	>
		<AlertDialog.Header>
			<AlertDialog.Title class="font-serif">다른 기기에서 이 초안을 먼저 고쳤어요</AlertDialog.Title
			>
			<AlertDialog.Description>어느 쪽 내용으로 이어 쓸까요?</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<Button
				variant="outline"
				class="min-h-11 border-input"
				disabled={resolving}
				onclick={overwrite}>지금 내용으로 덮어쓰기</Button
			>
			<Button class="min-h-11" disabled={resolving} onclick={loadTheirs}>그쪽 내용 불러오기</Button>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>

<!-- 초안이 없어짐 (DESIGN 4.3절). 쓴 내용을 먼저 버리지 않는다 -->
<AlertDialog.Root open={saveState === 'missing'}>
	<AlertDialog.Content
		class="rounded-md"
		escapeKeydownBehavior="ignore"
		interactOutsideBehavior="ignore"
	>
		<AlertDialog.Header>
			<AlertDialog.Title class="font-serif">이 초안을 찾을 수 없어요</AlertDialog.Title>
			<AlertDialog.Description>
				다른 기기에서 완료했거나 지웠을 수 있어요. 지금 쓴 내용은 화면에 그대로 있어요.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<Button
				variant="outline"
				class="min-h-11 border-input"
				disabled={resolving}
				onclick={() => {
					leaving = true;
					goto(resolve('/(app)'));
				}}>오늘로 돌아가기</Button
			>
			<Button class="min-h-11" disabled={resolving} onclick={startOver}>새 초안으로 저장</Button>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>

<!-- 저장에 실패한 채 나가기 -->
<AlertDialog.Root
	open={leaveTo !== null}
	onOpenChange={(o) => {
		if (!o) leaveTo = null;
	}}
>
	<AlertDialog.Content class="rounded-md">
		<AlertDialog.Header>
			<AlertDialog.Title class="font-serif">아직 저장되지 않은 내용이 있어요</AlertDialog.Title>
			<AlertDialog.Description>나가면 저장되지 않은 내용은 사라져요.</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel class="min-h-11 border-input">계속 쓰기</AlertDialog.Cancel>
			<Button variant="outline" class="min-h-11 border-input" onclick={leaveAnyway}
				>그냥 나가기</Button
			>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
