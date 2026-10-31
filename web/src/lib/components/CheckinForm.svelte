<script lang="ts">
	// S-03 체크인 입력 — 작성 · 수정 화면이 같이 쓴다 (DESIGN 4.2절).
	// 10초 안에 끝나야 한다(C1): 감정 칩 → (강도는 50에서 시작) → 저장. 메모 · 시각은 선택.
	import { untrack } from 'svelte';
	import type { components } from '#lib/api/schema.js';
	import { auth } from '#lib/auth.svelte.js';
	import EmotionPicker, { type Picked } from '#lib/components/EmotionPicker.svelte';
	import FormError from '#lib/components/FormError.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { fromLocalInput, toLocalInput } from '#lib/tz.js';

	type Emotion = components['schemas']['Emotion'];
	type CheckinInput = components['schemas']['CheckinInput'];

	let {
		emotions,
		initial,
		submitting = false,
		fields = {},
		message = '',
		onsubmit
	}: {
		emotions: Emotion[];
		/** 수정 모드일 때 기존 값 */
		initial?: { occurred_at: string; note: string; emotions: Picked[] };
		submitting?: boolean;
		fields?: Record<string, string>;
		message?: string;
		onsubmit: (input: CheckinInput) => void;
	} = $props();

	// (app) 틀은 로그인 정보가 있을 때만 화면을 그리므로 여기서는 늘 사용자 시간대가 있다
	const tz = $derived(auth.user?.timezone ?? 'Asia/Seoul');

	// 폼 값은 처음 한 번만 initial에서 가져온다 — 그 뒤로는 사용자가 고친 값이 주인이다.
	// 시각: 작성은 "지금"(저장하는 순간), [시각 바꾸기]를 누르면 직접 고른다. 수정은 기존 시각을 사용자 시간대 벽시계로 보여 준다.
	// untrack: initial이 나중에 바뀌어도 따라가지 않는다는 뜻을 분명히 (폼을 연 순간의 값만 쓴다)
	const start = untrack(() => initial);
	let picked = $state<Picked[]>(start?.emotions ?? []);
	let note = $state(start?.note ?? '');
	let customTime = $state(Boolean(start));
	let when = $state(
		start ? toLocalInput(start.occurred_at, auth.user?.timezone ?? 'Asia/Seoul') : ''
	);

	function pickTime() {
		customTime = true;
		when = toLocalInput(new Date(), tz);
	}

	function submit(event: SubmitEvent) {
		event.preventDefault();
		const body: CheckinInput = { note, emotions: picked };
		if (customTime && when) body.occurred_at = fromLocalInput(when, tz).toISOString();
		onsubmit(body);
	}
</script>

<form class="pb-24 md:pb-0" onsubmit={submit} novalidate>
	<section aria-labelledby="q-emotion">
		<h2 id="q-emotion" class="font-serif text-[26px] font-bold">지금 어떤 감정인가요?</h2>
		<p class="mt-1 mb-4 text-sm text-muted-foreground">여러 개 골라도 돼요</p>
		<EmotionPicker {emotions} bind:value={picked} />
		<div class="mt-2"><FormError message={fields.emotions} /></div>
	</section>

	<!-- 메모는 공책 줄처럼 밑줄 하나 (DESIGN 6장 "모양 D2" — 기록 칸) -->
	<div class="mt-6">
		<label for="note" class="text-sm text-muted-foreground">메모 (선택)</label>
		<input
			id="note"
			bind:value={note}
			maxlength="200"
			placeholder="한 줄로"
			autocomplete="off"
			aria-invalid={fields.note ? 'true' : undefined}
			class="mt-1 h-11 w-full border-0 border-b border-input bg-transparent px-0.5 outline-none placeholder:text-muted-foreground focus-visible:border-b-2 focus-visible:border-primary"
		/>
		<FormError message={fields.note} />
	</div>

	<div class="mt-6">
		<p class="text-sm text-muted-foreground" id="when-label">시각</p>
		{#if customTime}
			<input
				type="datetime-local"
				bind:value={when}
				aria-labelledby="when-label"
				aria-invalid={fields.occurred_at ? 'true' : undefined}
				class="mt-1 h-11 rounded-md border border-input bg-card px-3 tabular-nums"
			/>
		{:else}
			<div class="mt-1 flex items-center gap-3">
				<span>지금</span>
				<button
					type="button"
					class="min-h-11 text-sm text-accent-text underline-offset-4 hover:underline"
					onclick={pickTime}>시각 바꾸기</button
				>
			</div>
		{/if}
		<FormError message={fields.occurred_at} />
	</div>

	<div class="mt-4"><FormError {message} /></div>

	<!-- 저장: 모바일은 엄지가 닿는 화면 아래에 고정 (DESIGN 1장 원칙 1 · 4.2절) -->
	<div
		class="fixed inset-x-0 bottom-0 border-t bg-background px-5 pt-3 pb-[max(0.75rem,env(safe-area-inset-bottom))] md:static md:mt-8 md:border-0 md:p-0"
	>
		<Button
			type="submit"
			class="min-h-12 w-full text-base font-bold md:w-auto md:px-8"
			disabled={submitting || picked.length === 0}
		>
			{submitting ? '저장하는 중…' : '저장'}
		</Button>
	</div>
</form>
