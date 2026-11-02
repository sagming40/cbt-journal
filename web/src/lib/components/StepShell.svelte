<script lang="ts">
	// S-04의 틀 (DESIGN 4.3절, 5.2절): 머리(닫기 · 저장 상태) · 진행 막대 · 앞에서 쓴 답 · 제목 · 하단 이전/다음.
	// 작성과 수정이 같이 쓰고, 다른 점(저장 상태 · [나중에 이어 쓰기])은 snippet으로 받는다.
	import { tick, type Snippet } from 'svelte';
	import ChevronDown from '@lucide/svelte/icons/chevron-down';
	import ArrowLeft from '@lucide/svelte/icons/arrow-left';
	import type { components } from '#lib/api/schema.js';
	import AnswersPanel from '#lib/components/AnswersPanel.svelte';
	import FormError from '#lib/components/FormError.svelte';
	import { Button } from '#lib/components/ui/button/index.js';
	import { LAST_STEP, STEPS, type RecordForm } from '#lib/record-form.js';

	type Named = components['schemas']['Emotion'];

	let {
		heading,
		onback,
		step,
		form,
		emotions,
		distortions,
		message = '',
		nextLabel,
		nextDisabled,
		onprev,
		onnext,
		onjump,
		status,
		later,
		children
	}: {
		heading: string;
		/** 머리 ←. 어디로 가는지는 화면마다 다르다 (DESIGN 4.3절 "뒤로") */
		onback: () => void;
		step: number;
		form: RecordForm;
		emotions: Named[];
		distortions: Named[];
		message?: string;
		nextLabel: string;
		nextDisabled: boolean;
		onprev: () => void;
		onnext: () => void;
		onjump: (step: number) => void;
		status?: Snippet;
		later?: Snippet;
		children: Snippet;
	} = $props();

	const current = $derived(STEPS[step - 1]);
	// 1 ~ 3단계는 패널 없이 본문만(붙잡기). 4단계부터 앞에서 쓴 답이 보인다 (DESIGN 4.3절)
	const showAnswers = $derived(step >= 4);

	// 키보드 (DESIGN 4.3절 "키보드"): Ctrl(⌘) + Enter = 다음 · 완료. 한글 조합 중에는 마지막 글자가 확정 전이라 무시.
	// Esc는 일부러 아무 일도 하지 않는다 — 긴 글을 쓰는 화면이라 실수로 나가지 않게.
	function onkeydown(event: KeyboardEvent) {
		if (event.isComposing || event.defaultPrevented) return;
		if (event.key === 'Enter' && (event.ctrlKey || event.metaKey)) {
			event.preventDefault();
			if (!nextDisabled) onnext();
		}
	}

	// 단계를 옮기면 초점도 옮긴다: 마우스가 있는 기기는 바로 쓰게 입력칸으로,
	// 휴대폰은 키보드가 저절로 올라와 화면을 가리지 않게 제목으로 (화면 낭독기도 새 질문을 읽는다).
	let titleEl = $state<HTMLElement | null>(null);
	let bodyEl = $state<HTMLElement | null>(null);
	let lastStep = 0;
	$effect(() => {
		const s = step;
		if (lastStep === s) return;
		const first = lastStep === 0;
		lastStep = s;
		if (first) return;
		tick().then(() => {
			const box = bodyEl?.querySelector('textarea');
			if (box && window.matchMedia('(pointer: fine)').matches) box.focus();
			else titleEl?.focus({ preventScroll: false });
		});
	});
</script>

<svelte:window {onkeydown} />

<!-- 데스크톱(48rem 이상) 4 ~ 8단계: 본문 | 오른쪽 앞에서 쓴 답 (DESIGN 4.3절). 글줄은 40rem 안으로 -->
<div
	class={[
		'mx-auto pb-28 md:pb-0',
		showAnswers
			? 'md:grid md:grid-cols-[minmax(0,1fr)_16rem] md:items-start md:gap-10 lg:grid-cols-[minmax(0,40rem)_20rem]'
			: 'max-w-[40rem]'
	]}
>
	<div class="min-w-0">
		<header class="flex items-center gap-2">
			<button
				type="button"
				onclick={onback}
				class="-ml-2 flex size-11 shrink-0 items-center justify-center rounded-md hover:bg-muted"
				aria-label="뒤로"
			>
				<ArrowLeft class="size-5" aria-hidden="true" />
			</button>
			<h1 class="flex-1 font-serif text-lg font-bold">{heading}</h1>
			{@render status?.()}
		</header>

		<!-- 진행 막대: 칸 8개 + 숫자 -->
		<div class="mt-2 flex items-center gap-3 md:mt-3">
			<div class="flex flex-1 gap-1" aria-hidden="true">
				{#each STEPS as s, i (s.title)}
					<span
						class={[
							'h-1 flex-1 rounded-sm',
							i < step ? 'bg-primary dark:bg-accent-text' : 'bg-border'
						]}
					></span>
				{/each}
			</div>
			<span class="text-sm text-muted-foreground tabular-nums">
				<span class="sr-only">전체 {LAST_STEP}단계 중</span>{step} / {LAST_STEP}
			</span>
		</div>

		{#if showAnswers}
			<!-- 휴대폰: 본문 위 접힌 카드 (DESIGN 4.3절). 쓰다가 고개를 들면 앞의 답이 있다 -->
			<details class="group mt-4 rounded-md border bg-card px-3 md:hidden">
				<summary class="flex min-h-11 cursor-pointer list-none items-center justify-between">
					<span class="text-sm font-semibold">앞에서 쓴 답</span>
					<ChevronDown
						class="size-4 transition-transform group-open:rotate-180"
						aria-hidden="true"
					/>
				</summary>
				<div class="pb-2">
					<AnswersPanel {form} {step} {emotions} {distortions} {onjump} />
				</div>
			</details>
		{/if}

		<div class="mt-6 md:mt-8">
			<h2
				bind:this={titleEl}
				tabindex="-1"
				class="font-serif text-[26px] leading-snug font-bold outline-none md:text-[28px]"
			>
				{current.title}
			</h2>
			<p class="mt-1 mb-5 text-sm text-muted-foreground md:mb-6">{current.help}</p>
			<div bind:this={bodyEl}>
				{@render children()}
			</div>
			{#if message}<div class="mt-4"><FormError {message} /></div>{/if}
		</div>

		<!-- 이전 · 다음: 모바일은 엄지가 닿는 화면 아래 고정 (DESIGN 4.3절), 데스크톱은 본문 아래 -->
		<div
			class="fixed inset-x-0 bottom-0 z-10 border-t bg-background px-5 pt-3 pb-[max(0.75rem,env(safe-area-inset-bottom))] md:static md:mt-8 md:border-0 md:bg-transparent md:p-0"
		>
			<div class="flex items-center gap-2">
				<Button
					variant="outline"
					class="min-h-12 border-input px-5"
					disabled={step === 1}
					onclick={onprev}>이전</Button
				>
				<div class="flex-1"></div>
				{@render later?.()}
				<Button class="min-h-12 px-7 text-base font-bold" disabled={nextDisabled} onclick={onnext}>
					{nextLabel}
				</Button>
			</div>
			<p class="mt-2 hidden text-right text-xs text-muted-foreground md:block">
				Ctrl + Enter로 {nextLabel}
			</p>
		</div>
	</div>

	{#if showAnswers}
		<aside
			class="hidden rounded-lg border bg-card px-5 pt-4 pb-3 md:sticky md:top-8 md:block"
			aria-labelledby="answers-title"
		>
			<h3 id="answers-title" class="mb-2 font-serif text-base font-bold">앞에서 쓴 답</h3>
			<AnswersPanel {form} {step} {emotions} {distortions} {onjump} />
		</aside>
	{/if}
</div>
